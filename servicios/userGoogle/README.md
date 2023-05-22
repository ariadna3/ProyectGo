# UserGoogle

Las funciones principales de userGoogle es el CRUD de los usuarios mas algunas funciones extras que seran detalladas a continuación

- InsertUserITP: ingreso de un usuario
- GetUserITP: obtiene un usuario segun su email
- GetSelfUserITP: obtiene los datos del propio recurso que esta solicitando los datos mediante su token
- DeleteUserITP: elimina un usuario segun su email
- UpdateUserITP: Cambia el permiso y rol de un usuario

## Funciones internas

- validateGoogleJWT: a esta funcion se le envia un token y devuelve el email de google de la misma
- validacionDeUsuario: a esta funcion se le envia un bool "obligatorioAdministrador", un string "rolEsperado" y un token y devuelve un error y un string (un email si no hay error y el codigo de error si hubo algun problema) para validar que el usuario tiene o no los accesos solicitados (ser administrador y tener cierto rol)
- getGooglePublicKey: obtiene la clave publica de google para utilizar sus servicios OpenID Provider
