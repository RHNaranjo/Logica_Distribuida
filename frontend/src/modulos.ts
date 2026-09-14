export type Campo = {
  nombre: string;
  etiqueta: string;
  tipo: "text" | "area" | "numero";
  inicial: string;
  ayuda?: string;
};

export type Modulo = {
  id: string;
  nombre: string;
  descripcion: string;
  campos: Campo[];
  construirEntrada: (valores: Record<string, string>) => unknown;
};

// Quita espacios y líneas vacías de un texto 
function lineas(texto: string): string[] {
  return texto
    .split("\n")
    .map((linea) => linea.trim())
    .filter(Boolean);
}

// "q | MP | 1,3" => 
//  formula: "q"
//  regla: "MP" (mundo posible)
//  referencia: [1,3]
function parsearPaso(linea: string) {
  const partes = linea.split("|").map((parte) => parte.trim());

  const paso: { formula: string; regla: string; referencias =: number[] } = {
    formula: partes[0] ?? "",
    regla: (partes[1] ?? "").toUpperCase(),
  };

  if (partes[2]) {
    paso.referencias = partes[2]
      .split(",")
      .map((numero) => Number(numero.trim()))
      .filter((numero) => !Number.isNaN(numero));
  }

  return paso;
}

exponrt const modulos: Modulo[] = [
  {
    id: "tablas",
    nombre: "Tablas de verdad",
    descripcion: "Genera la tabla de verdad de lógica proposicional.",
    campos: [
      {
        nombre: "formula",
        etiqueta: "Fórmula",
        tipo: "texto",
        inicial: "p & (q -> r)",
        ayuda: "Operadores: ~ & | ->",
      },
    ],
    construirEntrada: (v) => ({ formula: v.formula }),
  },
  {
    id: "mundos",
    nombre: "Mundos posibles",
    descripcion: "Evalúa una fórmula proposicional a partir de relaciones entre mundos.",
    campos: [
      {
        nombre: "num_mundos", 
        etiqueta: "Número de mundos", 
        tipo: "numero", 
        inicial: "3",
      },
      {
        nombre: "valuaciones",
        etiqueta: "Valuaciones",
        tipo: "texto",
        inicial: "{0 | p, q}, {1 | ~p, q}, {2 | p}",
        ayuda: "Valor de cada mundo existente",
      },
      {
        nombre: "relaciones",
        etiqueta: "Relaciones",
        tipo: "texto",
        inicial: "<0,0>, <0,1>, <1,2>, <2,1>",
        ayuda: "Pares <origen,destino> separados por comas",
      },
      {
        nombre: "formula",
        etiqueta: "Fórmula modal",
        tipo: "texto",
        inicial: "[]p",
        ayuda: "[] es necesario, <> es posible",
      },
      {
        nombre: "mundo",
        etiqueta: "Mundo inicial",
        tipo: "numero",
        inicial: "0",
      },
    ],
    construirEntrada: (v) => ({
      num_mundos: v.num_mundos,
      relaciones: v.relaciones,
      valuaciones: v.valuaciones,
      formula: v.formula,
      mundo: Number(v.mundo),
    }),
  },
  {
    id: "deduccion",
    nombre: "Deducción natural",
    descripcion: "Verifica que cada paso de una demostración sea válido.",
    campos: [
      {
        nombre: "premisas",
        etiqueta: "Premisas",
        tipo: "area",
        inicial: "p -> q\nq -> r\np",
        ayuda: "Una proposición por línea",
      },
      {
        nombre: "conclusion",
        etiqueta: "Conclusión",
        tipo: "texto",
        inicial: "r",
      },
      {
        nombre: "pasos",
        etiqueta: "Pasos",
        tipo: "area",
        inicial: "p -> q | PREMISA\nq -> r | PREMISA\np | PREMISA\nq | MP | 1,3\nr | RM | 2,4",
        ayuda: "formula | REGLA | referencias (las referencias comienzan en el 1)",
      },
    ],
    construirEntrada: (v) => ({
      premisas: lineas(v.premisas),
      conclusion: v.conclusion.trim(),
      pasos: lineas(v.pasos).map(parsearPaso),
    }),
  },
];
