package lib

type botErrorType int

const (
	SomethingWentWrong botErrorType = 1000
	UserDontHaveAccess botErrorType = 1001
	UnCorrectUrl       botErrorType = 1002
)

var BotErrorToStr = map[botErrorType]string{
	SomethingWentWrong: "Что то пошло не так!",
	UserDontHaveAccess: "Доступ запрещен",
	UnCorrectUrl:       "Неправильная ссылка",
}
