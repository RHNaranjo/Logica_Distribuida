import { useEffect, useState } from "react";
import { pedir, type Estado } from "../api";

export function PanelEstado() {
  const [estadp, setEstado] = useState<Estado | null> (null);
  const [error, setError] = useState<string | null> (null);

  useEffect(() => {
    // Evita escribir el estado si se desmontó el componente y la petición sigue viva 
    // Se vuelve falsa si se cierra la página o el componente 
    let vivo = true;

    async function consultar() {
      try {
        const resp = await pedir<Estado>("/api/estado");
        if (!vivo) return;

        setEstado(resp.datos);
        setError(resp.ok ? null : `El middleware respondió ${resp.estado}`);
      } catch {
        if (!vivo) return;
        setError("No hay contacto con el middleware");
      }
    }

    consultar();
    const temporizador = setInterval(consultar, 3000);

    return () => {
      vivo = false;
      clearInterval(temporizador);
    };
  }, []);

  // Lista de grupos ordenados alfabéticamente
  const grupos = Object.entries(estado ?? {})
    .sort(([a], [b]) => a.localeCompare(b));

  // Cuántas unidades de la lista están sanas 
  const sanas = grupos
    .flatMap(([, lista]) => lista)
    .filter((i) => i.sana).length;

  // Número total en la lista 
  const total = grupos
    .flatMap(([, lista]) => lista).length;

  return (
    <section cassName="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <div className="mb-3 flex items-baseline justify-between">
        <h2 className="text-sm font-semibold tracking-wide text-slate-700 uppercase">
          Estado del sistema
        </h2>
        <span className="font-mono text-xs text-slate-500">
          {error ? "sin datos" : `${sanas}/${total} instancias sanas`}
        </span>
      </div>

      {error && (
        <p className="rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-sm text-red-700">
          {error}
        </p>
      )}

      {!error && total === 0 && (
        <p className="text-sm text-slate-500">Ninguna instancia se ha registrado aún.</p>
      )}

      <div className="grid gap-3 sm:grid-cols-3">
        {grupos.map(([modulo, instancias]) => (
          <div key={modulo} className="rounded border border-slate-200 p-3">
            <p className="mb-2 font-mono text-xs font-semibold text-slate-600">{modulo}</p>

            <ul className="space-y-1">
              {instancias.map((inst) => (
                <li key={inst.instancia} className="flex items-center gap-2 text-xs">
                  <span 
                    className={`inline-block h-2 w-2 shrink-0 rounded-full ${inst.sana ? "bg-emerald-500" : "bg-red-500"}`}
                    title={inst.sana ? "sana" : "no responde"}
                  />
                  <span className="font-mono text-slate-700">{inst.instancia}</span>
                  <span className="ml-auto font-mono text-slate-400">{inst.atendidas}</span>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </section>
  );
}
