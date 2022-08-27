package fileutil

import (
	"os"
	"strings"
)

// ファイル名に使えない・面倒なようなものを置換する
func SanitizeReplaceName(name string) string {
	rep := strings.NewReplacer(
		"?", "？",
		"!", "！",
		"*", "＊",
		"&", "＆",
		"\n", "",
		" ", "_",
		"　", "_",
		`\`, "_",
		"/", "_",
		":", "：",
		";", "；",
		"<", "＜",
		">", "＞",
		`"`, "_",
		`'`, "_",
		"|", "_",
		"(", "_",
		")", "）",
		"+", "＋",
	)
	return rep.Replace(name)
}

func MkdirAllIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0700)
	}
	return nil
}
