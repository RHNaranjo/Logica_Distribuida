-- Cada mundo tiene su propia base con el mismo esquema, pero distinto contenido
CREATE TABLE IF NOT EXISTS configuracion (
  id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  mundo_id INTEGER NOT NULL,
  actualizado TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Variables proposicionals del mundo. Pueden ser verdaderas o falsas 
CREATE TABLE IF NOT EXISTS variables (
  nombre TEXT PRIMARY KEY,
  valor BOOLEAN NOT NULL 
);

-- Mundos pueden ver la dirección de red del reso del proceso 
CREATE TABLE IF NOT EXISTS sucesores (
  mundo_id INTEGER PRIMARY KEY,
  url TEXT NOT NULL 
);
