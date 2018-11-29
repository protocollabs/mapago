package managementPlane

import "fmt"
import "github.com/monfron/mapago/control-plane/ctrl/shared"

func HandleMsmtStartReq(rec_ch chan<- shared.ChMsmt2Ctrl, msmt_start_req *shared.DataObj, clt_addr string) {
	fmt.Printf("\n\nrequest data struct is: % s", msmt_start_req)
	fmt.Printf("\n\nrequest data struct is: % s", msmt_start_req.Measurement.Type)
	// maybe we have to trim the port from the ip, we'll see that later
	fmt.Printf("\n\nclient addr is: % s", clt_addr)

}
