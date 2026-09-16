 // Lo que hay en cada teclado 
 export type Tecla = {
   muestra: string; // Lo que enseña el botón 
   inserta?: string; // Lo que escribe el botón. Dice 'muestra' si hace falta 
   accion?: "borrar" | "resolver";
   titulo?: string; // hover con el mouse 
 };

 export type Teclado = {
   campo: string; // en qué campo del formulario se escribe 
   filas: Tecla[][];
 };

 // Letras del teclado 
 const letras: Tecla[] = [
   { muestra: "A" },
   { muestra: "B" },
   { muestra: "C" },
   { muestra: "D" },
   { muestra: "E" },
   { muestra: "F" },
   { muestra: "X" },
   { muestra: "Y" },
   { muestra: "Z" },
 ];

 // Símbolos y conectores 
 const conectores: Tecla[] = [
   { muestra: "~", titulo: "negación" },
   { muestra: "&", inserta: " & ", titulo: "conjunción" },
   { muestra: "|", inserta: " | ", titulo: "disyunción" },
   { muestra: "→", inserta: " -> ", titulo: "condicional" },
   { muestra: "↔", inserta: " <-> ", titulo: "bicondicional" },
   { muestra: "(" },
   { muestra: ")" },
 ];


const acciones: Tecla[] = [
   { muestra: "⌫", accion: "borrar", titulo: "borrar el último carácter" },
   { muestra: "=", accion: "resolver", titulo: "resolver" },
 ];

 export const teclados: Record<string, Teclado> = {
   tablas: {
     campo: "formula",
     filas: [letras, conectores, acciones],
   },

   mundos: {
     campo: "formula",
     filas: [
       letras,
       conectores,
       // Símbolos de lógica modal 
       [
         { muestra: "□", inserta: "[]", titulo: "necesario" },
         { muestra: "◇", inserta: "<>", titulo: "posible" },
       ],
       acciones,
     ],
   },

   deduccion: {
     campo: "pasos",
     filas: [
       letras,
       conectores,
       [
         { muestra: "↵", inserta: "\n", titulo: "siguiente paso" },
         { muestra: "∀", titulo: "universal (sigue pendiente)" },
         { muestra: "∃", titulo: "existencial (sigue pendiente)" },
       ],
       acciones,
     ],
   },
 };
