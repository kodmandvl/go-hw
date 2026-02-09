//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})
}

// Добавим еще проверку на логику работы функции после моих изменений в коде.
// В т.ч. в 8-й строке у нас невалидный json,
// а в 7-й строке есть домен первого уровня "com", но нет "@" в email-е.
// В моей реализации я пропускаю строки, у которых невалидный json или нет "@" в email-е.
func TestGetDomainStat2(t *testing.T) {
	data := `{"Id":1,"Name":"Ivan Ivanov","Username":"0Oliver","Email":"aliquid_qui_ea@YANDEX.ru","Phone":"8-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@gOOgle.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@RAMbler.ru","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@example.com","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@yaNDeX.ru","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}
{"Id":6,"Name":"Eugen Fedorov","Username":"0Oliver","Email":"aliquid_qui_ea@MaiL.ru","Phone":"8-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":7,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynchbroWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":8,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com",Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39
{"Id":9,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":10,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@GooGLe.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find 'ru'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "ru")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"yandex.ru":  2,
			"rambler.ru": 1,
			"mail.ru":    1,
		}, result)
	})

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"google.com":  2,
			"example.com": 1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})
}
