// Conexión con el middleware
export type Respuesta<T> = {
  ok: boolean;
  estado: number;
  datos: T;
  instancia: string | null;
  recorrido: string[];
  peticion: string | null;
  ms: number;
};

export type Instancia = {
  instancia: string;
  url: string;
  sana: boolean;
  atendidas: number;
}

export type Estado = Record<string, Instancia[]>;


// Peticiones al middleware 
export async function pedir<T>(ruta: string, cuerpo?: unknown): Promise<Respuesta<T>> {
  const inicio = performance.now();

  const opciones: RequestInit =
    cuerpo === undefined
      ? { method: "GET" }
      : {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(cuerpo),
      };

  const resp = await fetch(ruta, opciones);
  const ms = Math.round(performance.now() - inicio);

  // Leer como texto porque 503 o 502 del MW vienen como JSON, pero otro error no necesariamente 
  const texto = await resp.text();

  let datos: unknown = texto;
  try {
    datos = JSON.parse(texto);
  } catch {
    // Se queda en texto plano al no tener formato JSON 
  }

  // X-Nodo viene repetido (middleware, loadbalancer, instancia)
  // La API de Headers une los valores repetidos con comas 
  const recorrido = (resp.headers.get("X-Nodo") ?? "")
    .split(",")
    .map((parte) => parte.trim())
    .filter(Boolean);

  return {
    ok: resp.ok,
    estado: resp.status,
    datos: datos as T,
    instancia: resp.headers.get("X-Instancia"),
    recorrido,
    peticion: resp.headers.get("X-Peticion"),
    ms,
  };
}
