package app

import (
	"fmt"
	"io"
	"museff/internal/common"
	"museff/internal/mis"
	"museff/internal/validator"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
type App struct {
	DB          *gorm.DB
	Echo        *echo.Echo
	InfoService *mis.MusicInfoService
}

type Config struct {
	Echo                *echo.Echo
	MusicInfoServiceUrl string
	DSN                 string
	LogLevel            log.Lvl
}

func NewApp(config Config) (a *App, err error) {
	a = &App{Echo: config.Echo}
	a.Echo.Logger.SetLevel(config.LogLevel)
	a.Echo.Logger.Info("Пытаемся подключиться к базе данных...")
	err = a.initDB(config.DSN)
	if err != nil {
		a.Echo.Logger.Fatalf("Ошибка подключения к базе данных: %s", err.Error())
		return
	}
	a.Echo.Logger.Info("Успешное подключение к базе данных")
	a.InfoService = mis.NewMusicInfoService(config.MusicInfoServiceUrl)

	a.Echo.Validator = validator.NewValidator()

	a.Echo.GET("/song/:id/:part", a.getSongLyricsHandler)
	a.Echo.GET("/song/:id", a.getSongById)
	a.Echo.GET("/song", a.searchSongsHandler)
	a.Echo.POST("/song", a.postSongHandler)
	a.Echo.DELETE("/song", a.deleteSongHandler)
	a.Echo.PATCH("/song", a.patchSongHandler)
	return
}

func (a *App) initDB(dsn string) error {
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return err
	}
	a.DB = db

	if err := db.AutoMigrate(&common.SongInfo{}); err != nil {
		return err
	}

	return nil
}

func (a *App) Start(address string) error {
	return a.Echo.Start(address)
}

// postSongHandler godoc
// @Summary      Добавить песню по группе и названию
// @Tags         Песни
// @Accept       json
// @Param        body body app.postSongHandler.req true  "Название и исполнитель"
// @Success      201
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /song [post]
func (a *App) postSongHandler(c echo.Context) error {
	type req struct {
		Group *string `json:"group" validate:"required" example:"Nothing More"`
		Song  *string `json:"song" validate:"required" example:"Angel song"`
	} // @name PostRequest

	var song req
	if err := c.Bind(&song); err != nil {
		c.Logger().Debug(err.Error())
		return c.NoContent(400)
	}

	err := c.Validate(&song)
	if err != nil {
		c.Logger().Debug(err.Error())
		return c.NoContent(400)
	}

	sng, err := a.InfoService.GetMusicInfo(*song.Group, *song.Song)
	if err != nil {
		c.Logger().Warn(err.Error())
		return c.NoContent(500)
	}

	t, err := time.Parse("02.01.2006", sng.ReleaseDate)
	if err != nil {
		c.Logger().Warn(err.Error())
		return c.NoContent(500)
	}

	dbsng := &common.SongInfo{
		Song:        *song.Song,
		Group:       *song.Group,
		ReleaseDate: t,
		Text:        sng.Text,
		Link:        sng.Link,
	}

	a.DB.Create(dbsng)
	return c.NoContent(201)
}

// deleteSongHandler godoc
// @Summary      Удалить песню по Id
// @Tags         Песни
// @Param        id   query int true  "ID песни"
// @Success      204
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /song [delete]
func (a *App) deleteSongHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil {
		c.Logger().Debug(err.Error())
		return c.String(400, "Bad Request")
	}

	var sng common.SongInfo
	tx := a.DB.Find(&sng, id)
	if tx.Error != nil || sng.ID == 0 {
		if tx.Error != nil {
			c.Logger().Debug(tx.Error.Error())
		} else {
			c.Logger().Debug("Песня не найдена")
		}
		return c.String(404, "Not Found")
	}

	tx = a.DB.Delete(&sng)
	if tx.Error != nil {
		c.Logger().Warn(tx.Error.Error())
		return c.String(500, "Internal Server Error")
	}
	return c.String(204, "Deleted")
}

// patchSongHandler godoc
// @Summary      Исправить данные песни
// @Tags         Песни
// @Param        id   query app.patchSongHandler.req true  "ID песни"
// @Success      204
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /song [patch]
func (a *App) patchSongHandler(c echo.Context) error {
	type req struct {
		ID          int `validate:"required"`
		Song        *string
		Group       *string
		ReleaseDate *time.Time
		Text        *string
		Link        *string
	}

	var si req
	if err := c.Bind(&si); err != nil {
		c.Logger().Debug(err.Error(), ":", si)
		return c.NoContent(400)
	}

	if err := c.Validate(&si); err != nil {
		c.Logger().Debug(err.Error(), ":", si)
		return c.NoContent(400)
	}

	var og common.SongInfo

	if tx := a.DB.Find(&og, si.ID); tx.Error != nil {
		c.Logger().Debug(tx.Error.Error(), si.ID)
		return c.NoContent(404)
	}

	if tx := a.DB.Model(&og).Where(si.ID).Updates(&si); tx.Error != nil {
		c.Logger().Warn(tx.Error.Error(), si.ID)
		return c.NoContent(500)
	}

	return c.NoContent(204)
}

// patchSongHandler godoc
// @Summary      Найти песню
// @Tags         Песни
// @Param        id   query app.searchSongsHandler.req true  "ID песни"
// @Success      200 {array} common.SongInfo
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /song [get]
func (a *App) searchSongsHandler(c echo.Context) error {
	type req struct {
		ID             *int    `query:"id"`
		Song           *string `query:"song"`                              // Название песни. Игнорирует регистр букв, для совпадения хватит части названия
		Group          *string `query:"group"`                             // Исполнитель. Игнорирует регистр букв, для совпадения хватит части имени/названия группы
		Text           *string `query:"text"`                              // Текст песни. Игнорирует регистр букв, для совпадения хватит части тектса
		ReleasedAfter  *string `query:"released_after"`                    // Даты должны быть в формате YYYY-MM-DD
		ReleasedBefore *string `query:"released_before"`                   // Даты должны быть в формате YYYY-MM-DD
		PageSize       int     `query:"page_size" validate:"gt=0,lte=100"` // Размер страницы. Не может быть больше 100 или меньше 0
		Page           int     `query:"page" validate:"gt=0"`              // Страница. Нумерация начинается с 1
	}

	si := req{Page: 1, PageSize: 10}
	if err := c.Bind(&si); err != nil {
		r, _ := io.ReadAll(c.Request().Body)
		c.Logger().Debugf("Не удалось спарсить тело запроса в структуру\nПолученная структура\nТело запроса", si, string(r))
		return c.String(400, "Bad Request")
	}

	if err := c.Validate(&si); err != nil {
		c.Logger().Debugf("Валидация провалилась\nСтруктура: %v", si)
		return c.String(400, "Bad Request")
	}

	sngs := make([]common.SongInfo, 0, si.PageSize)

	q := a.DB.Offset((si.Page - 1) * si.PageSize).Limit(si.PageSize)

	if si.ID != nil {
		q = q.Where(si.ID)
	}

	if si.Group != nil {
		q = q.Where("group ILIKE ?", fmt.Sprintf("%%%s%%", *si.Group))
	}

	if si.Song != nil {
		q = q.Where("song ILIKE ?", fmt.Sprintf("%%%s%%", *si.Song))
	}

	if si.Text != nil {
		q = q.Where("text ILIKE ?", fmt.Sprintf("%%%s%%", *si.Text))
	}

	if si.ReleasedBefore != nil {
		t, err := time.Parse(time.DateOnly, *si.ReleasedBefore)
		if err != nil {
			c.Logger().Debug("Ошибка парсинга времени: ", si.ReleasedBefore, "\n", err.Error())
			return c.String(400, "Bad Request")
		}
		q = q.Where("release_date < ?", t)
	}

	if si.ReleasedAfter != nil {
		t, err := time.Parse(time.DateOnly, *si.ReleasedAfter)
		if err != nil {
			c.Logger().Debug("Ошибка парсинга времени: ", si.ReleasedAfter, "\n", err.Error())
			return c.String(400, "Bad Request")
		}
		q = q.Where("release_date > ?", t)
	}

	tx := q.Find(&sngs)

	if tx.Error != nil {
		c.Logger().Warn(tx.Error.Error())
		return c.String(500, "Internal Server Error")
	}

	return c.JSON(200, sngs)
}

// getSongById godoc
// @Summary      Найти песню по ID
// @Tags         Песни
// @Param        id   path int true  "ID песни"
// @Success      204
// @Failure      404
// @Failure      500
// @Router       /song/{id} [get]
func (a *App) getSongById(c echo.Context) error {
	id := 0
	if err := echo.PathParamsBinder(c).Int("id", &id).BindError(); err != nil {
		c.Logger().Debug("Не удалось получить ID песни из пути")
		return c.NoContent(400)
	}
	var sng common.SongInfo
	tx := a.DB.Where(id).Find(&sng)
	if tx.Error != nil {
		c.Logger().Warnf("Ошибка базы данных при попытке найти песню по ID.\nid: %d", id)
		return c.NoContent(500)
	}
	if sng.ID == 0 {
		c.Logger().Debugf("Не найдена песня с ID: %d", id)
		return c.NoContent(404)
	}
	return c.JSON(200, &sng)
}

// getSongLyricsHandler godoc
// @Summary      Найти фрагмент слов песни по ID и номеру фрагмента
// @Tags         Песни
// @Param        id   path int true  "ID песни"
// @Param        part path int true  "Номер фрагмента"
// @Success      204
// @Failure      404
// @Failure      500
// @Router       /song/{id}/{part} [get]
func (a *App) getSongLyricsHandler(c echo.Context) error {
	var (
		id   int
		part int
		res  struct {
			SongID     int
			Part       int
			TotalParts int
			Text       string
		}
	)
	err := echo.PathParamsBinder(c).Int("id", &id).Int("part", &part).BindError()
	if err != nil {
		c.Logger().Debug(err.Error(), c.Request().URL.Path, id, part)
		return c.String(400, "Bad Request")
	}

	var sng common.SongInfo
	tx := a.DB.Find(&sng, id)
	if tx.Error != nil || sng.ID == 0 {
		if tx.Error != nil {
			c.Logger().Debug(tx.Error.Error())
		} else {
			c.Logger().Debug("Песня не найдена")
		}
		return c.String(404, "Not Found")
	}

	parts := strings.Split(sng.Text, "\\n\\n")

	if len(parts) <= part {
		c.Logger().Debugf("Запрос на больше частей чем существует: %d/%d(ID: %d)", part, len(parts), id)
		return c.JSON(404, struct {
			SongID     int
			TotalParts int
		}{SongID: id, TotalParts: len(parts)})
	}

	res = struct {
		SongID     int
		Part       int
		TotalParts int
		Text       string
	}{
		SongID:     id,
		Part:       part,
		TotalParts: len(parts),
		Text:       parts[part],
	}

	return c.JSON(200, &res)
}
