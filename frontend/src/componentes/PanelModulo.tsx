import { useEffect, useState } from "react";

import { pedir, type Respuesta } from "../api";
import type { Modulo } from "../modulos";

// arranca los campos con sus valores de ejemplo 
function valoresIniciales(modulo: Modulo): Record<string, string> {
  return Object.fromEntries(modulo.campos.map((campo) => [campo.nombre, campo.inicial]));
}

export function PanelModulo({ modulo }: { modulo: Modulo }) {
  const [valores, setValores] = useState(() => valoresIniciales(modulo));
  const [respuesta, setRespuesta] = useState<Respuesta<unknown> | null>(null);
  const [fallo, setFallo] = useState<string | null>(null);
  const [enviado, setEnviado] = useState(false);

  // Al cambiar de pestaña se reinicia el formulario 
  useEffect(() => {
    setValores(valoresIniciales(modulo));
    setRespuesta(null);
    setFallo(null);
  }, [modulo]);

  // Envía valores al módulo y guarda resultados 
  async function resolver() {
    // Mostrar solicitud y borrar errores anteriores
    setEnviado(true);
    setFallo(null);

    try {
      // Convierte los valores en entradas válidas 
      const entrada = modulo.construirEntrada(valores);

      // Solicita resolución al endpoint
      setRespuesta(await pedir(`/api/${modulo.id}/resolver`, entrada));
    } catch (err) {
      // No muestra resultados a una petición fallida 
      setRespuesta(null);

      // Guarda mensaje legible para el usuario 
      setFallo(err instanceof Error ? err.message : "[ERROR] No se pudo contactar al middleware");
    } finally {
      // Oculta estado al finalizar 
      setEnviado(false);
    }
  }

  return (
    <section className="grid gap-4 lg:grid-cols-2">
      <div className="rounded-lg border border-black bg-white p-4">
        <p className="mb-4 text-sm text-slate-600">{modulo.descripcion}</p>

        <div className="space-y-3">
          {modulo.campos.map((campo) => (
            <label key={campo.nombre} className="block">
              <span className="mb-1 block text-xs font-semibold text-slate-700">
                {campo.etiqueta}
              </span>

              {campo.tipo === "area" ? (
                <textarea 
                  rows={5}
                  className="w-full rounded border border-slate-300 px-2 py-1.5 font-mono text-sm focus:border-slate-500 focus:outline-none"
                  value={valores[campo.nombre]}
                  onChange={(e) => setValores({ ...valores, [campo.nombre]: e.target.value })}
                />
              ) : (
                <input 
                  type={campo.tipo === "numero" ? "number" : "text"}
                  className="w-full rounded border border-slate-300 px-2 py-1.5 font-mono text-sm focus:border-slate-500 focus:outline-none"
                  value={valores[campo.nombre]}
                  onChange={(e) => setValores({ ...valores, [campo.nombre]: e.target.value })}
                />
              )}

              {campo.ayuda && (
                <span className="mt-1 block text-xs text-slate-500">{campo.ayuda}</span>
              )}
            </label>
          ))}
        </div>

        <button
          onClick={resolver}
          disabled={enviado}
          className="mt-4 rounded bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-700 disabled:opacity-50"
        >
          {enviado ? "Resolviendo..." : "Resolver"}
        </button>
      </div>

      <div className="rounded-lg border border-black bg-white p-4">
        <h2 className="mb-3 text-sm font-semibold tracking-wide text-slate-700 uppercase">Respuesta</h2>

        {fallo && (
          <p className="rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{fallo}</p>
        )}

        {!fallo && !respuesta && (
          <p className="text-sm text-slate-500">Todavía no has resuelto nada.</p>
        )}

        {respuesta && (
          <>
            <dl className="mb-3 grid grid-cols-2 gap-2 text-xs">
              <Dato etiqueta="Código" valor={String(respuesta.estado)} />
              <Dato etiqueta="Tiempo" valor={`${respuesta.ms} ms`} />
              <Dato etiqueta="Instancia" valor={respuesta.instancia ?? "-"} />
              <Dato etiqueta="Petición" valor={respuesta.peticion ?? "-"} />
            </dl>

            <p className="mb-3 flex flex-wrap items-center gap-1 text-xs">
              {respuesta.recorrido.map((nodo, i) => (
                <span key={`${nodo}-${i}`} className="flex items-center gap-1">
                  {i > 0 && <span className="text-slate-400">→</span>}
                  <span className="rounded bg-slate-100 px-2 py-0.5 font-mono text-slate-700">
                    {nodo}
                  </span>
                </span>
              ))}
            </p>

            <pre className="max-h-80 overflow-auto rounded bg-slate-900 p-3 font-mono text-xs text-slate-100">
              {JSON.stringify(respuesta.datos, null, 2)}
            </pre>
          </>
        )}
      </div>
    </section>
  );
}

function Dato({ etiqueta, valor }: { etiqueta: string; valor: string }) {
  return (
    <div>
      <dt className="text-slate-500">{etiqueta}</dt>
      <dd className="font-mono text-slate-800">{valor}</dd>
    </div>
  );
}
