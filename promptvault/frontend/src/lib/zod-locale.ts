// Глобально включаем русскую локаль Zod v4 — иначе при ошибке валидации
// без custom-message пользователь видит сырое английское «Invalid input:
// expected string, received undefined». После z.config(ru()) такие сообщения
// автоматически рендерятся на русском («Введите …» / «Должно быть строкой»
// и т.п. со склонениями).
//
// Импорт делается из main.tsx ДО первой формы.
import { z } from "zod"
import { ru } from "zod/v4/locales"

z.config(ru())
