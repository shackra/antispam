# Ershin: un bot de Telegram para el chat GNU Emacs en Español

Hice este bot en una tarde con algo de ayuda de Grok (sugerencias de su parte nada más). Elimina a las cuentas de spammers que suelen entrar al chat.

La idea es muy sencilla, pregunta al usuario sobre qué comando corresponde cierta combinación de teclas del `global-keymap`, aquellos que contestan incorrectamente son expulsados de por vida y sus mensajes eliminados. La pregunta y su respuesta es recogida bajo demanda de una instancia de Emacs (iniciada con `-q`) por lo que nada es almacenado en el código fuente.

## Cosas pendientes

Habilitar sub-prefijos para aumentar la cantidad de opciones posibles a escoger por el bot, por ahora hay aproximadamente unas 28 que corresponde a los atajos a nivel global.
