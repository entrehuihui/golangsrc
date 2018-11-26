package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_AesEncrypt(t *testing.T) {
	Convey("test AesEncrypt", t, func() {
		val := "123456"
		val_en, err := AesEncrypt([]byte(val))
		So(err, ShouldBeNil)
		val_de, err := AesDecrypt(val_en)
		So(err, ShouldBeNil)
		So(string(val_de), ShouldEqual, val)
		So(val_en, ShouldEqual, "Qg3iB+W+g91FQQlrVkfRtw==")
	})
}
