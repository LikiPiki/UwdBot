## Плагины

Каждый функионал бота разбит на плагины, например за игру РПГ отвечает wars, а за показ профиля profiler. Все они реализуют интерфейс ``Plugin``, его можно найти в файле plugin.go.

Каждый плагин состоит из двух файлов:
- ``{plugin_name}-handlers.go`` - хендлеры для обработки ботом, которые удовлетворяют интерфейсу ``Plugin``
- ``{plugin_name}.go`` - логика плагина

> __Как быстро создать плагин?__
Устанавливаем impl если вы это еще не сделали!
```
go get -u github.com/josharian/impl
```

Теперь быстро сгенериуем функции заглушки для интерфейса
```
impl 'p *Profiler' plug.Plugin
```
__Выхлоп:__
```go
func (p *Profiler) Init(s *sender.Sender) {
        panic("not implemented")
}

func (p *Profiler) HandleMessages(msg *tgbotapi.Message) {
        panic("not implemented")
}

func (p *Profiler) HandleCommands(msg *tgbotapi.Message, command string) {
        panic("not implemented")
}

func (p *Profiler) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *data.User) {
        panic("not implemented")
}

func (p *Profiler) HandleCallbackQuery(update *tgbotapi.Update) {
        panic("not implemented")
}

func (p *Profiler) HandleAdminCommands(msg *tgbotapi.Message) {
        panic("not implemented")
}

func (p *Profiler) GetRegisteredCommands() []string {
        panic("not implemented")
}
```

***Внимание!!!***
- Обязательно уберите везде ``panic("not implemented")`` и вставьте в ваш файл с плагином.

- Создайте структуру ``Profiler`` в данном случае
```go
type Profiler struct {
	c *sender.Sender // Для отправки сообщений
}

func (p *Profiler) Init(s *sender.Sender) {
	p.c = s
}
```
- Добавьте плагин в список инициализации в файле main.go
```go
// Зарегистрируйте свой плагин тут
plugins := pl.Plugins{
    &pl.Base{},
    &pl.Wars{},
    &pl.Minigames{},
    &pl.Profiler{}, // Например вот так
}
```