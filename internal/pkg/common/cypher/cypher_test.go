package cypher_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/cypher"
	"github.com/Buzzvil/go-test/test"
)

const API_AES_KEY = "4i3jfkdie0998754"

func TestCypher(t *testing.T) {
	params := map[string]interface{}{
		"a": 1234,
		"b": "5678",
	}
	data := cypher.EncryptAesBase64Dict(params, API_AES_KEY, API_AES_KEY, true)
	t.Log(fmt.Sprintf("TestCypher.EncryptAesBase64Dict() - %v", data))

	params2 := cypher.DecryptAesBase64Dict(data, API_AES_KEY, API_AES_KEY, true)
	t.Log(fmt.Sprintf("TestCypher.DecryptAesBase64Dict() - %v", params2))
	test.AssertEqual(t, len(params), len(params2), "TestCypher - DecryptAesBase64Dict")

	params = map[string]interface{}{
		"c":   259458,
		"d":   1092,
		"udt": "1b66a654cd0d4c4fb471f4fb02b65015",
		"u":   238793895727606,
		"i":   "069e6a97-b341-43a0-b9fc-df2556f06a25",
		"cou": "KR",
		"sex": "M",
	}
	data = cypher.EncryptAesBase64Dict(params, API_AES_KEY, API_AES_KEY, true)
	t.Log(fmt.Sprintf("TestCypher.EncryptAesBase64Dict() - %v", data))
	params2 = cypher.DecryptAesBase64Dict(data, API_AES_KEY, API_AES_KEY, true)
	t.Log(fmt.Sprintf("TestCypher.DecryptAesBase64Dict() - %v", params2))
	data = "9y40zU-99Gt4EnrcJ1c5FAuR_uG07-F-iKJ3f68ok5iFUQxGroDJ6rVcFr6gr84r5xhxCMkHLPDalSI0JBmdfMtmU5_tMsTJV76v6ULRZ1IGd4XY2MVcy-4L_2CbbtjA3CX-vBqRchHL2y7JK3xW0w=="

	params3 := map[string]interface{}{
		"cou":      "KR",
		"ended_at": 1504537200,
		"oid":      1,
		"sex":      "M",
		"t":        1504509049,
		"tz":       "Asia/Seoul",
		"yob":      1985,
	}

	text3 := cypher.EncryptAesBase64Dict(params3, API_AES_KEY, API_AES_KEY, true)
	text3 = "aN9jXxl5eYceFPqGwREeWcCvjAgXhRyMns08ooPg2BjKlNk5SBVTDHq2rfkvh95VCsu2kXzZoCIaigCf9BXrcDUJ-QJ8nezqCAx1Q467a-xYXUvGornIpXWWst0k0cAp0_MrYThLZnbiy0ha0CPJ6A=="
	params3 = cypher.DecryptAesBase64Dict(text3, "ejuvbej14828vjdp", "ejuvbej14828vjdp", true)
	t.Log(fmt.Sprintf("TestCypher.DecryptAesBase64Dict() - data: %v, map: %v", params3, text3))
}

func TestBase64(t *testing.T) {
	t.Log(fmt.Sprintf("TestBase64() - %v", base64.RawURLEncoding.EncodeToString([]byte("abcd&="))))
}

/*
func TestCypher(t *testing.T) {
	params := map[string]interface{}{
		"u":   1234,
		"c":   5678,
		"i":   "abcd",
		"d":   5678,
		"udt": "efgh",
		"cou": "kor",
		"yob": 1984,
		"sex": "M",
	}
	data := EncryptAesDict(params, API_AES_KEY, API_AES_KEY, false)
	fmt.Println(data)
	test.AssertEqual(t, data != "", true, "TestCypher")
}
*/
