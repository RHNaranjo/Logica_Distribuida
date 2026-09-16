import type { Tecla, Teclado as DatosTeclado } from "../teclados";

type Props = {
  teclado: DatosTeclado;
  onInsertar: (texto: string) => void;
  onBorrar: () => void;
  onResolver: () => void;
};

export function Teclado({ teclado, onInsertar, onBorrar, onResolver }: Props) {
  function click(tecla: Tecla) {
    if (tecla.accion === "borrar") {
      onBorrar();
      return;
    }

    if (tecla.accion === "resolver") {
      onResolver();
      return;
    }

    // Insertar lo que se muestre
    onInsertar(tecla.inserta ?? tecla.muestra);
  }

  return (
    <div className="space-y-2">
      {teclado.filas.map((fila, i) => (
        <div key={i} className="flex flex-wrap gap-1">
          {fila.map((tecla) => (
            <button
              key={tecla.muestra}
              type="button"
              title={tecla.titulo}
              onClick={() => click(tecla)}
              className={`min-w-9 rounded border border-black px-2 py-1 font-mono text-sm ${
                tecla.accion === "resolver"
                  ? "bg-slate-900 text-white hover:bg-slate-700"
                  : "bg-white hover:bg-slate-100"
              }`}
            >
              {tecla.muestra}
            </button>
          ))}
        </div>
      ))}
    </div>
  );
}
