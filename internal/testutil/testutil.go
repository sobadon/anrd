package testutil

import "github.com/pkg/errors"

// github.com/pkg/errors の errors.As に nil も扱えるようにしたもの
// 第一引数に gotErr
// 第二引数に wantErr が期待されている
func ErrorsAs(err error, target interface{}) bool {
	// nil と nil の比較のため
	if err == target {
		return true
	}

	// errors.As の target が nil だとダメなのでそれを防ぐ
	if err != nil && target == nil {
		return false
	}

	return errors.As(err, &target)
}
