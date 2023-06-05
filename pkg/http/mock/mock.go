package mock

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xX32/bNhD+Vwiuj3Ysy4pjGQiGDi3aYCsWdA0woMkMWjzJbCVSJU9Z3NT/+0BS/iFL",
	"TtM06V76ktAUed/xu++Ox1uaqKJUEiQaOr2lJllAwdzwpdZK20GpVQkaBbhpsNNuxMEkWpQolKRTeiHn",
	"qpIcOMmFQaJSUjJcGIKKsDwnQl6zXHCGwAlD1GJeIRjao7gsgU6pmn+ABOmqRwswhmXQRnhdFUwSDYyz",
	"eQ7EOULWq3sUblhR5tbWmccinyrQS1IyzYotkEEtZGaB/MQ+ijs2cd92bdb+zw5YW/Wohk+V0MDp9D3d",
	"elUbUiW96jirAztny1wxfoBqO3imIaVT+stgG6xBHamBD9O+A35rF+SZNMhkAm24eZV8BGwz4ueJZAUQ",
	"IYlBpfcZ90v6CAa7mE6URJA4y0FmuGgj1N+J/068swewjuOTOOzRVOmCoQsMjqMtqJAIGehd1O44rzHt",
	"17sRKcINDsqcCdl5OA1W1TPmqNv4ZaXeR1FA1x5AlrVdQpYdciEOY0gDdjKaM+DJcD6HyXgcRJNJMowY",
	"hKMuEMHbEGbBhmTBzJrkJsjJBI7jNAp4GIXxPIwmk1GUxCcwTCd8GMdBFEXx8LgT7CMs22g7QIcONrCa",
	"cX+O8AaJTb163IVSADLOkLnsWNu4pQtg3K6wUS1SYSMglCRQMJFvM30nCYz43CEJO3u3u+EoiMP4Xuoz",
	"yLDqqJN+nsA1yCb7XkdfOOSA8CVR5bKLgark36i2vcogOPXRqlXYEHDD/uYMvXVpqInbS61Wfu/E6a4K",
	"5MgRCIX5Wo3b1KxtLJnWbGl/vwEu2MESuhbLfawba65kGcxQIcvttv2w7nG5s7hHD573XMNfIpNnsqyw",
	"7SFLvDBuKciqsFZfvXxHe/T1y+cvaI+eX7zbMbrVwbZYt4vLTSk0zAohGxI5LNXO3K3zwF7iNhkKIYV6",
	"cObu8VafeUdX1oU7uPuu+NY2LnTecuRrMbN7WpiVzjt437NsF3UZvrg4e9Emu5LiUwVEcJAoUgHa9k+i",
	"1iXRkCjNG+RP5nE65hPWh+GY9yM2CfrzOEj7Ix6GCUt5OE4YtWpGBG0R/nkf9GPWT69uJ6vLy3l/8zNa",
	"HRxfXs43P4fh6llXPbowoN/U2X4mU/WgtuJ/uyJ/zK2VM4OzQnEbWf7Ast3IlE3pblru1FsjPp0V6I5S",
	"8vSV4bHu9FYDXhu9R43Zpei7Ck0rF+5ZbuwyUecOCnQcFvZaq6VHDOhrkViRXIM2PgrDo8A6r0qQrBR0",
	"SkdHwdHI5/zCuTxwNuwo68q/V4DErdgUGtMjBphOFqSOA1kgltPBgKuCCent/eqbgtO6a7isgiAcr+k+",
	"+rv/vPjctyT034HBWWVAn3LBMqkMisSvVmlqAE8D/ysXhcDToXskgXZRPuPeP3e3U8uhKZU0PhJhENh/",
	"dc/hrtCyzEXidg4+GH+Z+ph8LWKN5sHFocnRn79bjqMgejTExouvA/EtYKWlISIl9i7AJfmXGSIVktQ+",
	"r607x49IwDe4U0m4KSGxD3j/9FZJUmmrbLsNWWZ84m11S69sQ1V1aM93mcTtl2sZWq2QncRtyuG88nLY",
	"TTLq0wsM/qb48tFIadfMVTOTUVewekJZdlWkn+p8fHWuenRQajAic06VynRoNQNpdQjErZQZsa1du1jJ",
	"dcv49o8nkmXjHfGDFbnXh/8U4xOIcdWj9qIHbb/Wzwxa38C5Sli+UAanoyAI6OpqY+aWug662THQ1dXq",
	"vwAAAP//8rj5mtsVAAA=",
}

func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}

	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	var buf bytes.Buffer

	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()

	return func() ([]byte, error) {
		return data, err
	}
}

func pathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = pathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]

		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)

			return nil, err1
		}

		return getSpec()
	}

	var specData []byte
	specData, err = rawSpec()

	if err != nil {
		return
	}

	swagger, err = loader.LoadFromData(specData)

	if err != nil {
		return
	}

	return
}
