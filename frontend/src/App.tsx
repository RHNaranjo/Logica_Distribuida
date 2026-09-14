import { useState } from "react";

import { PanelEstado } from "./componentes/PanelEstado";
import { PanelModulo } from "./componentes/PanelModulo";
import { modulos } from "./modulos";

export default function App() {
  const [activo, setActivo] = useState(modulos[0].id);
  const modulo = modulos.find((m) => m.id === activo) ?? modulos[0];

  return ( 
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto max-w-5xl px-6 py-4">
          <h1 className="text-lg font-semibold">Lógica Distribuida</h1>
          <p className="text-sm text-slate-500">Tres servicios replicados detrás de un loadbalancer y un middleware</p>
        </div>
      </header>

      <main className="mx-auto max-w-5xl space-y-4 px-6 py-6">
        <PanelEstado />

        <nav className="flex gap-1 border-b border-slate-200">
          {modulos.map((m) => (
            <button
              key={m.id}
              onClick={() => setActivo(m.id)}
              className={`-mb-px border-b-2 px-4 py-2 text-sm font-medium ${
                m.id === activo
                  ? "border-slate-900 text-slate-900"
                  : "border-transparent text-slate-500 hover:text-slate-800"
              }`}
            >
              {m.nombre}
            </button>
          ))}
        </nav>

        <PanelModulo modulo={modulo} />
      </main>
    </div>
  );
}
