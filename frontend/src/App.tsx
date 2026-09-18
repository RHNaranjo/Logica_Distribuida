import { useState } from "react";

import { Navbar, type Seccion } from "./componentes/Navbar";
import { PanelEstado } from "./componentes/PanelEstado";
import { PanelModulo } from "./componentes/PanelModulo";
import { modulos } from "./modulos";

// Secciones de los backends
const secciones: Seccion[] = [
  { id: "inicio", nombre: "Inicio" },
  ...modulos.map((m) => ({ id: m.id, nombre: m.nombre })),
];

export default function App() {
  const [activa, setActiva] = useState("inicio");

  const modulo = modulos.find((m) => m.id === activa);

  return ( 
    <div className="min-h-screen bg-white text-slate-900">
      <Navbar secciones={secciones} activa={activa} onCambiar={setActiva} />

      <main className="mx-auto max-w-5xl space-y-4 px-6 py-6">
        {activa === "inicio" ? (
          <>
            <p className="text-sm text-slate-600">
              Conexión con el middleware y estado de instancias.
            </p>
            <PanelEstado />
          </>
        ) : modulo ? (
          <PanelModulo modulo={modulo} />
        ) : null}
      </main>
    </div>
  );
}
