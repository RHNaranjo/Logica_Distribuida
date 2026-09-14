CREATE TABLE IF NOT EXISTS historial (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  modulo TEXT NOT NULL,
  instancia, TEXT NOT NULL,
  entrada JSONB NOT NULL,
  resultado JSONB,
  error TEXT,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS historial_modulo_creado ON historial (modulo, creado_en DESC);
