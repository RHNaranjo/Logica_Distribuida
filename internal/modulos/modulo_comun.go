package modulos

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"logica/internal/historial"
	"logica/internal/rpc"
)

const (
	tiempoLimite     = 10 * time.Second
	limitePorOmision = 20
	limiteMaximo     = 200
)

// Esto sí cambia entre los distintos módulos
type Resolutor func(entrada json.RawMessage) (any, error)

// Retorno si no hay lógica
var ErrNoImplementado = errors.New("[ERROR] Este módulo no ha sido implementado")

type Controlador struct {
	rpc.UnimplementedModuloServer

	modulo    string
	instancia string
	repo      *historial.Repo
	resolver  Resolutor
}

func NuevoControlador(modulo, instancia string, repo *historial.Repo, resolver Resolutor) *Controlador {
	return &Controlador{modulo: modulo, instancia: instancia, repo: repo, resolver: resolver}
}

// Resolver la entrada propia del módulo
func (c *Controlador) Resolver(ctx context.Context, peticion *structpb.Struct) (*structpb.Struct, error) {
	// Convertir la entrada a JSON
	entrada, err := protojson.Marshal(peticion)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "no se pudo leer la entrada")
	}

	resultado, errResolver := c.resolver(entrada)

	// Anotar los intentos que funcionaron o no
	c.anotar(ctx, entrada, resultado, errResolver)

	if errors.Is(errResolver, ErrNoImplementado) {
		return nil, status.Error(codes.Unimplemented, errResolver.Error())
	}

	if errResolver != nil {
		return nil, status.Error(codes.InvalidArgument, errResolver.Error())
	}

	// Devolver el resultado como una Struct
	salida := &structpb.Struct{}

	if err := deJSON(resultado, salida); err != nil {
		return nil, status.Error(codes.Internal, "no se pudo armar la respuesta")
	}

	return salida, nil
}

func (c *Controlador) Historial(ctx context.Context, peticion *rpc.PeticionHistorial) (*structpb.ListValue, error) {
	limite := int(peticion.GetLimite())

	// Usar el límite por omisión si no se mandó
	if limite == 0 {
		limite = limitePorOmision
	}

	if limite < 0 || limite > limiteMaximo {
		return nil, status.Error(codes.InvalidArgument, "el límite debe estar entre 1 y 200")
	}

	// Limitar el tiempo de la consulta
	ctx, cancel := context.WithTimeout(ctx, tiempoLimite)
	defer cancel()

	registros, err := c.repo.Listar(ctx, c.modulo, limite)
	if err != nil {
		log.Printf("[ERROR] historial de la %s: %v", c.modulo, err)

		// Ocultar detalles
		return nil, status.Error(codes.Internal, "no se pudo leer el historial ")
	}

	// Devolver los registros como lista
	lista := &structpb.ListValue{}
	if err := deJSON(registros, lista); err != nil {
		return nil, status.Error(codes.Internal, "no se pudo armar el historial")
	}

	return lista, nil
}

// Guardar la base
func (c *Controlador) anotar(ctx context.Context, entrada []byte, resultado any, errResolver error) {
	// Crear el registro
	reg := historial.Registro{
		Modulo:    c.modulo,
		Instancia: c.instancia,
		Entrada:   entrada,
	}

	// Guardar el error por si acaso
	if errResolver != nil {
		reg.Error = errResolver.Error()
	} else if resultado != nil {
		// Conservar historial
		crudo, err := json.Marshal(resultado)

		if err != nil {
			log.Printf("[AVISO] No se pudo serializar el resultado de la instancia %s: %v", c.instancia, err)
		} else {
			// Guardar por so acaso
			reg.Resultado = crudo
		}
	}

	// Limitar el tiempo
	ctxAnotar, cancel := context.WithTimeout(ctx, tiempoLimite)
	defer cancel()

	if err := c.repo.Guardar(ctxAnotar, reg); err != nil {
		log.Printf("[ERROR] %v", err)
	}
}

// Pasar un valor de Go a un mensaje protobuf por medio de un JSON
func deJSON(valor any, destino proto.Message) error {
	crudo, err := json.Marshal(valor)
	if err != nil {
		return err
	}

	return protojson.Unmarshal(crudo, destino)
}
