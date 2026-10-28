// Lo que devuelve el módulo de tablas
type Renglon = { valores: boolean[]; parciales: boolean[]; resultado: boolean };

export type Tabla = {
  formula: string;
  variables: string[];
  columnas: string[];
  renglones: Renglon[];
  clasificacion: string;
};

// Mostrar V o F en vez de true o false
function letra(valor: boolean) {
  return valor ? "V" : "F";
}

export function TablaVerdad({ tabla }: { tabla: Tabla }) {
  // La última columna es la fórmula completa
  const ultima = tabla.columnas.length - 1;

  return (
    <div>
      <p className="mb-2 text-sm text-slate-600">
        Clasificación: <span className="font-semibold text-slate-900">{tabla.clasificacion}</span>
      </p>

      <div className="max-h-96 overflow-auto rounded border border-slate-300">
        <table className="w-full border-collapse text-center font-mono text-sm">
          <thead className="sticky top-0 bg-slate-100">
            <tr>
              {tabla.variables.map((v) => (
                <th key={v} className="border-b border-slate-300 px-3 py-1.5">{v}</th>
              ))}
              {tabla.columnas.map((c, j) => (
                <th
                  key={c}
                  className={`border-b border-l border-slate-300 px-3 py-1.5 whitespace-nowrap ${j === ultima ? "bg-slate-200" : ""}`}
                >
                  {c}
                </th>
              ))}
            </tr>
          </thead>

          <tbody>
            {tabla.renglones.map((r, i) => (
              <tr key={i} className="odd:bg-white even:bg-slate-50">
                {r.valores.map((v, k) => (
                  <td key={k} className="px-3 py-1">{letra(v)}</td>
                ))}
                {r.parciales.map((v, j) => (
                  <td
                    key={j}
                    className={`border-l border-slate-200 px-3 py-1 ${j === ultima ? `font-semibold ${v ? "text-emerald-700" : "text-red-600"}` : ""}`}
                  >
                    {letra(v)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
