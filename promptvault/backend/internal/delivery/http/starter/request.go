package starter

// CompleteRequest — body POST /api/starter/complete.
//
// Install может быть пустым массивом — означает «Пропустить wizard», маркируем
// юзера как онбординг-прошедшего без создания промптов.
type CompleteRequest struct {
	Install []string `json:"install" validate:"max=100,dive,max=100"`
}
