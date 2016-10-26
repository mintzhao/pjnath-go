/*
Copyright Mojing Inc. 2016 All Rights Reserved.
Written by mint.zhao.chiu@gmail.com. github.com: https://www.github.com/mintzhao

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package pjnath

// #cgo pkg-config: libpjproject
// #include "turn.h"
// #include "auth.h"
// int wapper_pj_AF_INET()
// {
//   return pj_AF_INET();
// }
import "C"
import (
	"fmt"
	"net"
	"unsafe"
)

const (
	REALM string = "pjsip.org"
)

func init() {
	status := PjInit()
	if !status.Success() {
		panic(fmt.Sprintf("pj_init() error: %v", status))
	}

	PjUtilInit()
	PjNathInit()
	PjTurnAuthInit(REALM)
}

type ICEListener struct {
	g_cp       *C.pj_caching_pool
	srv        *C.pj_turn_srv
	p_listener *C.pj_turn_listener
	addr       net.Addr
}

// Listen start a ice network, `net` must be *tcp* or *udp*
func Listen(net, host string, port int) (net.Listener, error) {
	if net != "tcp" && net != "udp" {
		return nil, fmt.Errorf("unsupported net protocal")
	}

	listener := &ICEListener{
		g_cp:       (*C.pj_caching_pool)(unsafe.Pointer(C.malloc(C.sizeof_pj_caching_pool))),
		srv:        (*C.pj_turn_srv)(unsafe.Pointer(C.malloc(C.sizeof_pj_turn_srv))),
		p_listener: (*C.pj_turn_listener)(unsafe.Pointer(C.malloc(C.sizeof_pj_turn_listener))),
		addr: &ICEAddr{
			host: host,
			port: port,
		},
	}
	C.pj_caching_pool_init(listener.g_cp, nil, 0)

	// create turn server
	if status := C.pj_turn_srv_create(&listener.g_cp.factory, &listener.srv); status != C.PJ_SUCCESS {
		return nil, fmt.Errorf("Error creating server: %v", status)
	}

	// create udp listener
	if status := C.pj_turn_listener_create_udp(listener.srv, C.wapper_pj_AF_INET(), parse2Pjstr(host), C.uint(port), 1, 0, &listener.p_listener); status != C.PJ_SUCCESS {
		return nil, fmt.Errorf("Error creating UDP listener: %v", status)
	}

	// create tcp listener
	if net == "tcp" {
		if status := C.pj_turn_listener_create_tcp(listener.srv, C.wapper_pj_AF_INET(), parse2Pjstr(host), C.uint(port), 1, 0, &listener.p_listener); status != C.PJ_SUCCESS {
			return nil, fmt.Errorf("Error creating listener: %v", status)
		}
	}

	if status := C.pj_turn_srv_add_listener(listener.srv, listener.p_listener); status != C.PJ_SUCCESS {
		return nil, fmt.Errorf("Error adding listener: %v", status)
	}
	fmt.Println("server is running")

	return listener, nil
}

func (l *ICEListener) Accept() (net.Conn, error) {
	return nil, nil
}

func (l *ICEListener) Close() error {
	C.pj_turn_srv_destroy(l.srv)
	C.pj_caching_pool_destroy(l.g_cp)
	C.pj_shutdown()

	fmt.Println("server is closing")
	return nil
}

func (l *ICEListener) Addr() net.Addr {
	return l.addr
}

// new implement of net.Addr
type ICEAddr struct {
	host string
	port int
}

func (a *ICEAddr) Network() string {
	return "ICE"
}

func (a *ICEAddr) String() string {
	return fmt.Sprintf("%s://%s:%d", a.Network(), a.host, a.port)
}

// wapper for C.pj_turn_auth_init
func PjTurnAuthInit(realM string) {
	C.pj_turn_auth_init(C.CString(realM))
}

// wapper for C.pjlib_util_init
func PjUtilInit() {
	C.pjlib_util_init()
}

// wapper for C.pjnath_init
func PjNathInit() {
	C.pjnath_init()
}

// wapper for C.pj_init
func PjInit() *PJStatus {
	return &PJStatus{
		status: C.pj_init(),
	}
}

// wapper for C.pj_status_t
type PJStatus struct {
	status C.pj_status_t
}

func (s *PJStatus) Success() bool {
	return s.status == C.PJ_SUCCESS
}

func (s *PJStatus) String() string {
	return fmt.Sprintf("status: %v", s.status)
}

// parse2Pjstr parse golang string to pj's string
func parse2Pjstr(str string) *C.pj_str_t {
	pjstr := (*C.pj_str_t)(unsafe.Pointer(C.malloc(C.sizeof_pj_str_t)))
	pjstr.ptr = C.CString(str)
	pjstr.slen = C.pj_ssize_t(C.strlen(pjstr.ptr))

	return pjstr
}

// SetLogger
func SetLogger(level int) {
	C.pj_log_set_level(C.int(level))
}
