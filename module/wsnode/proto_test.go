package wsnode

import (
	"fmt"
	"testing"
	"time"
)

func TestProto(t *testing.T) {
	req := &Req{
		Sn:  fmt.Sprint(time.Now().UnixNano()),
		Uri: "testing",
		Ver: "0.1",
		Param: map[string]interface{}{
			"testing": "testing",
		},
	}

	reqS := req.Serial()
	if _, err := ParseReq(reqS); err != nil {
		t.Fatal(err)
	}

	resp := &Resp{
		Sn:   fmt.Sprint(time.Now().UnixNano()),
		Code: "200",
		Uri:  "testing",
		Data: map[string]interface{}{
			"testing": "testing",
		},
	}

	respS := resp.Serial()
	if _, err := ParseResp(respS); err != nil {
		t.Fatal(err)
	}

}
