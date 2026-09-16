export type Seccion = {
  id: string;
  nombre: string;
};

type Props = {
  secciones: Seccion[];
  activa: string;
  onCambiar: (id: string) => void;
};

export function Navbar({ secciones, activa, onCambiar }: Props) {
  return (
    <header className="border-b border-black bg-white">
      <div className="mx-auto max-w-5xl px-6">
        <div className="py-4">
          <h1 className="text-lg font-semibold">Lógica Distribuida</h1>
          <p className="text-sm text-slate-500">
            Con: 1 MW, 1 LB y 3 instancias para 3 backends.
          </p>
        </div>

        <nav className="flex flex-wrap gap-1">
          {secciones.map((seccion) => (
            <button
              key={seccion.id}
              type="button"
              onClick={() => onCambiar(seccion.id)}
              className={`-mb-px border-b-2 px-4 py-2 text-sm font-medium ${
                seccion.id === activa 
                  ? "border-slate-900 text-slate-900"
                  : "border-transparent text-slate-500 hover:text-slate-800"
              }`}
            >
              {seccion.nombre}
            </button>
          ))}
        </nav>
      </div>
    </header>
  );
}
