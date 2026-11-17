package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"slices"

	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/jwt"
)

type RPCUsecase struct {
	rpc    repository.RPCRepository
	auth   repository.AuthRepository
	tokens *Tokens
}

func NewRPCUsecase(rpc repository.RPCRepository, auth repository.AuthRepository, tokens *Tokens) *RPCUsecase {
	return &RPCUsecase{rpc: rpc, auth: auth, tokens: tokens}
}

func (u *RPCUsecase) Execute(ctx context.Context, fn, token string, params map[string]json.RawMessage) (json.RawMessage, *Failure) {
	params = maps.Clone(params)
	if params == nil {
		params = map[string]json.RawMessage{}
	}

	var name string
	if i := slices.Index(publicRPCs, fn); i >= 0 {
		name = publicRPCs[i]
	} else {
		claims, role, f := u.authorize(ctx, token)
		if f != nil {
			return nil, f
		}
		i := slices.Index(rolePermissions[role], fn)
		if i < 0 {
			return nil, fail(http.StatusForbidden, fmt.Sprintf("Forbidden: Peran Anda (%s) tidak diizinkan memanggil fungsi ini", role))
		}
		name = rolePermissions[role][i]
		injectIdentity(name, role, claims, params)
	}

	raw, err := u.rpc.Call(ctx, name, params)
	if err != nil {
		msg, ok := dbErrorMessage(err)
		if !ok {
			return nil, fail(http.StatusInternalServerError, "Internal server error")
		}
		client := msg
		if systemErrorRe.MatchString(msg) {
			client = "Terjadi kesalahan sistem database: " + msg
		}
		return nil, failDetails(http.StatusBadRequest, client, msg)
	}
	out, err := stripCredentials(raw)
	if err != nil {
		return nil, fail(http.StatusInternalServerError, "Internal server error")
	}
	return out, nil
}

func (u *RPCUsecase) authorize(ctx context.Context, token string) (claims map[string]any, role string, f *Failure) {
	if token == "" {
		return nil, "", fail(http.StatusUnauthorized, "Unauthorized: Missing session token")
	}
	claims, err := u.tokens.Verify(token)
	if err != nil {
		return nil, "", fail(http.StatusUnauthorized, "Unauthorized: Invalid or expired token")
	}
	if isRevoked(ctx, u.auth, claims) {
		return nil, "", fail(http.StatusUnauthorized, "Unauthorized: Session telah berakhir. Silakan login kembali.")
	}
	// rpc.ts memanggil toLowerCase() di dalam try: role non-string melempar dan berakhir 401, role kosong lanjut ke 403.
	rawRole, isString := claims["role"].(string)
	if !isString && claims["role"] != nil {
		return nil, "", fail(http.StatusUnauthorized, "Unauthorized: Invalid or expired token")
	}
	role = jwt.NormalizeRole(rawRole)
	if _, ok := rolePermissions[role]; !ok {
		return nil, "", fail(http.StatusForbidden, "Forbidden: Peran pengguna tidak valid")
	}
	return claims, role, nil
}

// C-2: token yang sudah logout ditolak. Error lookup dianggap belum dicabut, sama dengan
// kode lama yang mengabaikan error RPC is_token_revoked.
func isRevoked(ctx context.Context, auth repository.AuthRepository, claims map[string]any) bool {
	if !truthy(claims["jti"]) {
		return false
	}
	revoked, err := auth.IsTokenRevoked(ctx, claimText(claims["jti"]))
	return err == nil && revoked
}

// Parameter identitas ditimpa dari JWT supaya klien tidak bisa menyamar sebagai user lain.
// Klaim id yang tidak ada menghapus key, sama dengan nilai undefined di JSON.stringify;
// klaim id null melewati injeksi karena kode lama memeriksa userId !== null.
func injectIdentity(fn, role string, claims map[string]any, params map[string]json.RawMessage) {
	id, hasID := claims["id"]
	if hasID && id == nil {
		return
	}
	var idJSON json.RawMessage
	if hasID {
		b, err := json.Marshal(id)
		if err != nil {
			return
		}
		idJSON = b
	}
	has := func(key string) bool {
		_, ok := params[key]
		return ok
	}
	set := func(key string) {
		if hasID {
			params[key] = idJSON
		} else {
			delete(params, key)
		}
	}
	setIfPresent := func(keys ...string) {
		for _, k := range keys {
			if has(k) {
				set(k)
			}
		}
	}

	setIfPresent("p_viewer_id")
	if has("p_caller_id") || fn == "request_reschedule_secure" {
		set("p_caller_id")
	}
	setIfPresent("p_student_id")

	isStaff := role == "asisten" || role == "koordinator"
	switch {
	case !isStaff:
		setIfPresent("p_user_id", "p_target_user_id", "p_target_id")
	case slices.Contains(staffSelfRPCs, fn):
		setIfPresent("p_user_id")
	}

	// p_target_id dari klien tetap dipakai supaya staf bisa melihat profil orang lain; DB yang memeriksa.
	if fn == "get_user_profile" {
		set("p_caller_id")
		if !has("p_target_id") {
			set("p_target_id")
		}
	}
	// Hanya koordinator yang boleh mengganti password orang lain.
	if fn == "update_password" && role != "koordinator" {
		setIfPresent("p_target_id")
	}
}
