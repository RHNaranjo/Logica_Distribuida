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

// El servidor del frontend entrega env vars al navegador en config.json 
let baseGuardada: string | null = null;

// obtener la base (si es que no se tiene aún )
async function base(): Promise<string> {
  if (baseGuardada !== null) {
    return baseGuardada;
  }

  try {
    const resp = await fetch("/config.json");

    if (resp.ok) {
      const cfg = (await resp.json()) as { middleware?: string };

      if (cfg.middleware) {
        baseGuardada = cfg.middleware;
        return baseGuardada;
      }
    }
  } catch {
    // Se usa el respaldo 
  }

  baseGuardada = `http://${location.hostname}:8080`;
  return baseGuardada;
}

// Peticiones al middleware 
export async function pedir<T>(ruta: string, cuerpo?: unknown): Promise<Respuesta<T>> {
  const inicio = performance.now();

  const url = (await base()) + ruta;

  const opciones: RequestInit =
    cuerpo === undefined
      ? { method: "GET" }
      : {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(cuerpo),
      };

  const resp = await fetch(url, opciones);
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
