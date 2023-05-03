# Novedades

Dentro de las novedades se encuentran el manejo de novedades y cecos.

## Novedades

Las funciones principales de las novedades son el CRUD de las mismas mas algunas funciones extras que seran detalladas a continuación

- InsertNovedad: ingreso de las novedades
- GetNovedades: obtiene una novedad segun su id
- GetNovedadFiltro: obtiene novedades segun diversos filtros
- GetNovedadesAll: obtiene todas las novedades
- DeleteNovedad: elimina una novedad segun su id
- UpdateEstadoNovedades: modifica el estado y motivo de una novedad
- UpdateMotivoNovedades: modifica el motivo de una novedad
- GetFiles: obtiene una novedad segun la posición del array o segun su nombre
- GetTipoNovedad: obtiene una lista de tipo de novedades disponibles

## Cecos

Los centros de costos (cecos) tienen como funciones el ingreso y consulta de datos. Se supone que estos son permanentes y se obtienen directamente de la base de datos de bejerman por lo que no existe forma alguna de eliminarlos o modificarlos por medio de esta api. 

- InsertCecos: ingreso de centros de costos
- GetCecosAll: obtiene todos los cecos
- GetCecos: obtiene un ceco segun el legajo
- GetCecosFiltro: obtiene los cecos que cunplan con unos filtros
