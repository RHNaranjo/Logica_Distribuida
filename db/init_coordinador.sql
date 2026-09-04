-- Cada nodo guarda sólo un fragmento 
-- Lo que el CRUD edita y se reparte al guardar 

CREATE TABLE IF NOT EXISTS modelos (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  nombre TEXT NOT NULL,
  num_mundos INTEGER NOT NULL CHECK (num_mundos > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- La llave primaria impide que un par se guarde dos veces 
CREATE TABLE IF NOT EXISTS relaciones (
  modelo_id BIGINT NOT NULL REFERENCES modelos(id) ON DELETE CASCADE,
  origen INTEGER NOT NULL,
  destino INTEGER NOT NULL,
  PRIMARY KEY (modelo_id, origen, destino)
);

-- Una fila por variable 
-- valor = true -> usuario escribió p 
-- valor = false -> usuario escribió ~p 
CREATE TABLE IF NOT EXISTS valuaciones (
  modelo_id BIGINT NOT NULL REFERENCES modelos(id) ON DELETE CASCADE,
  mundo INTEGER NOT NULL,
  variable TEXT NOT NULL,
  valor BOOLEAN NOT NULL,
  PRIMARY KEY (modelo_id, mundo, variable)
;
