package jwtauth

import (
	"encoding/json"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/pkg/errors"
)

// Test keys, these are a key pair for testing, do not edit.
const testPrivateKey = `{"use":"sig","kty":"RSA","kid":"test","alg":"RS256","n":"y7cjmFqCb8B3OeZ9-OWDviKE0LCqEGoVP51R2uBDYxxVHoNyKfw-3Y4Trdxvr70IPdi8Fv-ubkQmUqZPG9V2VTsYrCaG6bdn0hXEspaIxIWiiKHPVZwUmP2p1Y0W7G0eyzxhGqYwkTxHUtmdiPmLtHLoww5SWsMNtSlx3emUVgjmK6jzvl7WlU-IqOmsKmNnwmNJT0N35-skirjpbIemEnTqolMgebR6BMgcMih6Etwb_mMtXx_Cir5MmSIlE94K3XXlQK4YlSaJWgyVGmClF3oQGdnU6oFtkOR_2qMCmLBaVoUJHM8TYoypeQo7eXkeMY5RcQTnAmrM1yERQDda7Q","e":"AQAB","d":"DfW317nkFFphETOtUEASHhZpeY-Rp9xNJnXWZSRXGdSYNKYXTa8-c5pH2PMxKB7REMPoZ78Pqfi7k5BX_XVMTZzmHO3q8tshnzDksMRGDQmHUMud1KUEeFNxrvOjLYJwyMaTdOsMivdRf-jvtbI8E5qIPs2dcSDKsK5tYiKeiqKj_l6eIe4fpUGmLW-9I-4bEeE8FSlkkWuGHXOOIkIrDI-jjQFGeLqhBOG1om2zFRK-VWCcvZN-Y7i8x90aUVk0ytK895QlTsCF6C-v0jUG_JwkYYhQJsA_xq6v1sPVOvHgaPEsDS7JruIs2xBgriFDdtFyLF_FqWlS0va4g2e_FQ","p":"_uPxlbYqvW7OySaNaNbTd92A5zNC0OOiyNhpNqqdo2_-dDAcx4GgJF32INPvb0TnkEnss8Yem8V7f4Jh70aJBCk_E5dqxv7yoATdJPNrynC8zi9ShRxK9JDYEJO3GRMaplM2yW1hGltHqTYltVaUdkjiOekyXMct2g1narW3xfc","q":"zJoqJHwh_tm0BqTaF65alRbbxeYbkrv_a8XTOkOUFuLWOJ5JEwa5M4jowO6mLHL9Eq5PQBIt5pKxL-XKvs-yRl_gaiA4uP7tP09HuI1JCUzlN49l-6pkYw7XypHmJ7mCIYabblm01cMovhE610eMnAwwrl7XJ7hNowZJyJ8eXTs","dp":"clgfkGHgWorTCTnaRiXZR_R-VzmPx9XWuPMcgAGaJi-fns_WmUl1ZdshBykMSIZIu1nubdd687Zr6I-9D3I9FTjLqyZKKGtGvLJx4pmwyWg5yuU_x6chmQVBaG5bvUvssKCz-ziuTvyT8TzxPaBRiZ64nfIXPbe8wg0xT5Wlk5E","dq":"DjegOgzOJ_FkyxlldkPNU5LVDrlgrR-XLhv_A4tynOyCSkjldwb-k5At7EopDemnoEawhxk8S0tiOJvVLNEt6Cn9ZCJ9Do3YWo_dwfs_WgAX5XZ3dbdvZlq_r_nXbmc7nazG3DIrmLcI-7wixJbaNHW8ZEF-3em2q19ifShhacU","qi":"YLzIIexkkjwBfueKG-d7T5E3rEmOy9NbT84o4vMy805Q2ZFGASL4qwIr-wwVnQcaEC1UijaMlWvUUgbQwg-F0qrymeGO9ZOmmz7VYSJL6h-DniYs72hBnOn90XMJJp5vPNcYYK7DRP5ffobKhDG8yYc6sA-BH1eZPj4wY3qi1hU"}`
const testJWKS = `{"keys":[{"use":"sig","kty":"RSA","kid":"test","alg":"RS256","n":"y7cjmFqCb8B3OeZ9-OWDviKE0LCqEGoVP51R2uBDYxxVHoNyKfw-3Y4Trdxvr70IPdi8Fv-ubkQmUqZPG9V2VTsYrCaG6bdn0hXEspaIxIWiiKHPVZwUmP2p1Y0W7G0eyzxhGqYwkTxHUtmdiPmLtHLoww5SWsMNtSlx3emUVgjmK6jzvl7WlU-IqOmsKmNnwmNJT0N35-skirjpbIemEnTqolMgebR6BMgcMih6Etwb_mMtXx_Cir5MmSIlE94K3XXlQK4YlSaJWgyVGmClF3oQGdnU6oFtkOR_2qMCmLBaVoUJHM8TYoypeQo7eXkeMY5RcQTnAmrM1yERQDda7Q","e":"AQAB"}]}`

// Public key from the above jwks, for testing verification.
var testPublicKey = getTestPublicKey()

// Key with no corresponding public key.
const testUntrustedPrivateKey = `{"use":"sig","kty":"RSA","kid":"0c0ee7a7-3cbb-4906-b265-325795398373","alg":"RS256","n":"0AX50ilbpiwtb88qA7DpqmB_UeTAxinZfp1GeFNmoJc7gCIJRZ1nmFTyuG4yRTZsrF5zsmfwKfpmNbw4Zn7eH4_WpLo_xVFV1cTd73zpNkUZS1A0VCchmk5MdQX6bn7bNaWhs1MiA4zmRdrPjDmaStNtI65sq78fdNsDJNGdeUgsHzXK-ux58Zc95Y_NMUCKpj9NxLSPAtB-ztGta-eyno4kv8LUT_qEkXBoCl7rd_PChmgiZfMRy7md_mHHhU4BrA6IJ1dWUIO4XjWNc9g96VNRw-tTKh4nJSWaHD_7p1DtvPy0FU37o5F6qqVgGpw2QYiFN6H_Jcdbv_FISZ311Q","e":"AQAB","d":"VyvZ50argDQNCkiOu6M8F8h-Mgwg-Cl7WcKAUFXqsKsPJP-eUQzH557ZY43SBQzsd0LRStahXoXupN_t5o2NeT5mXPsrU_1kccMgmYYHmFVWZygB9a28OBRNl9BchEcmhxGUdDgHDePSbz1lMcbFsEtu9b_XDBV4Ehjg9WHdkn948PFIAXGyxuKmU1MfCVCniNDwfMsspS_XM45uKBRxAK0shtiME6T6KI3xiupBEkxpk7TVrUoMk2WakqxR8hQHet1etDo4O6wiqgJUYgdHqtRXVGzXUDwobClsQBTf9Xd3T9WP20kyt3hpI9HulmhgGGfswDYbemQvL9JURvQqoQ","p":"-duYrkY1G6CywhIfdW7IKktZhypf8BZxVxcw0rLeshqLr9xQatKNc1sUj4_-G1ebRTGkcdDrozS0u7O4PkbOtlthTB7afN7dQH9Os1ru0GeetKaJUjcjFYIyRKsgJAZdBMWEbL8EIxKe6YlN_uUEaxWBNgdcPsfDjqBjSwYJP0k","q":"1SMba7WWARW7w8xqXUFeVUkw_T1LY2a91zZEIoEBrwsiY_9a7n09Njk8-xTNX3NcD5iCnKRbObRCZc7YEF3m7UnwRTf9z1S7Kx_WGOmWy1i1eqoJDL_nMsgBZXFDVrUVYtW8bDSeQzyarUeICOvGeMCMyMm5_FtLy9F0BLDRJi0","dp":"Zi-EYwn9oF35ndtxmEKFhJ6qb9hJwlQ7aGXoptNWtrqalILjNL0F8r62SvyV7TLIIuVpns7WADqHDBk1aerlkbkPsuUPcHBpRn6KflnbP8qRIsrVcJVyONK1olXmYDVmB5SMUzlQBNQRv-tSxcN-KhlybdlWxapHdWZtFXrTf1k","dq":"HaI79cXRjWUQLjEFuOGV1BXREeSrzq5CRuHspz94lHXf2jdu1SnkkN10dRR3WYYYjrKNtmnDpUpC0RTpRZ1ItkVJetZGG8WUIHLUubIAnVVAJkXt7C_iXVUhnJEa47tZtdwxznmiZ4bNmroPV-4wMinTaTdi_ItVBomgr-ZFriE","qi":"lMsb7BDq6h4NErBVOpOwWQiAFF6vC5sj54Qs1OGqglIzcjLgsJAzUBnuh4LLDCA_EkyAjqnh_JioMh32yX-5P3FSEELgI8A048UWFkWwnV5PnlriNQYK0iG78Am7pMdqdsV8Jbp18TTK6rFJPtwfftkJ270h8Vj5CG1uP23XsDQ"}`

// This private key is the same as the untrusted private key, but claims to be from a trusted source.
const testMaliciousPrivateKey = `{"use":"sig","kty":"RSA","kid":"test","alg":"RS256","n":"0AX50ilbpiwtb88qA7DpqmB_UeTAxinZfp1GeFNmoJc7gCIJRZ1nmFTyuG4yRTZsrF5zsmfwKfpmNbw4Zn7eH4_WpLo_xVFV1cTd73zpNkUZS1A0VCchmk5MdQX6bn7bNaWhs1MiA4zmRdrPjDmaStNtI65sq78fdNsDJNGdeUgsHzXK-ux58Zc95Y_NMUCKpj9NxLSPAtB-ztGta-eyno4kv8LUT_qEkXBoCl7rd_PChmgiZfMRy7md_mHHhU4BrA6IJ1dWUIO4XjWNc9g96VNRw-tTKh4nJSWaHD_7p1DtvPy0FU37o5F6qqVgGpw2QYiFN6H_Jcdbv_FISZ311Q","e":"AQAB","d":"VyvZ50argDQNCkiOu6M8F8h-Mgwg-Cl7WcKAUFXqsKsPJP-eUQzH557ZY43SBQzsd0LRStahXoXupN_t5o2NeT5mXPsrU_1kccMgmYYHmFVWZygB9a28OBRNl9BchEcmhxGUdDgHDePSbz1lMcbFsEtu9b_XDBV4Ehjg9WHdkn948PFIAXGyxuKmU1MfCVCniNDwfMsspS_XM45uKBRxAK0shtiME6T6KI3xiupBEkxpk7TVrUoMk2WakqxR8hQHet1etDo4O6wiqgJUYgdHqtRXVGzXUDwobClsQBTf9Xd3T9WP20kyt3hpI9HulmhgGGfswDYbemQvL9JURvQqoQ","p":"-duYrkY1G6CywhIfdW7IKktZhypf8BZxVxcw0rLeshqLr9xQatKNc1sUj4_-G1ebRTGkcdDrozS0u7O4PkbOtlthTB7afN7dQH9Os1ru0GeetKaJUjcjFYIyRKsgJAZdBMWEbL8EIxKe6YlN_uUEaxWBNgdcPsfDjqBjSwYJP0k","q":"1SMba7WWARW7w8xqXUFeVUkw_T1LY2a91zZEIoEBrwsiY_9a7n09Njk8-xTNX3NcD5iCnKRbObRCZc7YEF3m7UnwRTf9z1S7Kx_WGOmWy1i1eqoJDL_nMsgBZXFDVrUVYtW8bDSeQzyarUeICOvGeMCMyMm5_FtLy9F0BLDRJi0","dp":"Zi-EYwn9oF35ndtxmEKFhJ6qb9hJwlQ7aGXoptNWtrqalILjNL0F8r62SvyV7TLIIuVpns7WADqHDBk1aerlkbkPsuUPcHBpRn6KflnbP8qRIsrVcJVyONK1olXmYDVmB5SMUzlQBNQRv-tSxcN-KhlybdlWxapHdWZtFXrTf1k","dq":"HaI79cXRjWUQLjEFuOGV1BXREeSrzq5CRuHspz94lHXf2jdu1SnkkN10dRR3WYYYjrKNtmnDpUpC0RTpRZ1ItkVJetZGG8WUIHLUubIAnVVAJkXt7C_iXVUhnJEa47tZtdwxznmiZ4bNmroPV-4wMinTaTdi_ItVBomgr-ZFriE","qi":"lMsb7BDq6h4NErBVOpOwWQiAFF6vC5sj54Qs1OGqglIzcjLgsJAzUBnuh4LLDCA_EkyAjqnh_JioMh32yX-5P3FSEELgI8A048UWFkWwnV5PnlriNQYK0iG78Am7pMdqdsV8Jbp18TTK6rFJPtwfftkJ270h8Vj5CG1uP23XsDQ"}`

// Test signers to create jwts for testing.
// Uses private keys above.
var testSigner = createTestSigner(testPrivateKey)
var testUntrustedSigner = createTestSigner(testUntrustedPrivateKey)
var testMaliciousSigner = createTestSigner(testMaliciousPrivateKey)

type testVerifier struct{}

func (t testVerifier) Verify(token *jwt.JSONWebToken, claims ...interface{}) error {
	return token.Claims(testPublicKey, claims...)
}

func createTestSigner(key string) jose.Signer {
	var privKey jose.JSONWebKey
	if err := json.Unmarshal([]byte(key), &privKey); err != nil {
		panic(err)
	}
	if privKey.IsPublic() {
		panic(errors.New("Private key is not private"))
	}
	sig := jose.SigningKey{
		Algorithm: jose.SignatureAlgorithm(privKey.Algorithm),
		Key:       privKey,
	}
	signer, err := jose.NewSigner(sig, nil)
	if err != nil {
		panic(err)
	}
	return signer
}

func getTestPublicKey() *jose.JSONWebKey {
	var verifyJWKS jose.JSONWebKeySet
	if err := json.Unmarshal([]byte(testJWKS), &verifyJWKS); err != nil {
		panic(err)
	}
	return &verifyJWKS.Keys[0]
}

func issueTestJWT() string {
	claims := Claims{
		"iss": "test",
	}
	res, _ := jwt.Signed(testSigner).Claims(claims).Serialize()
	return res
}

func issueExpiredJWT() string {
	claims := Claims{
		"iss": "test",
		"exp": jwt.NewNumericDate(time.Now().Add(-5 * time.Second)),
	}
	res, _ := jwt.Signed(testSigner).Claims(claims).Serialize()
	return res
}

func issueUntrustedTestJWT() string {
	claims := Claims{
		"iss": "untrusted",
	}
	res, _ := jwt.Signed(testUntrustedSigner).Claims(claims).Serialize()
	return res
}

func issueMaliciousTestJWT() string {
	claims := Claims{
		"iss": "test",
	}
	res, _ := jwt.Signed(testMaliciousSigner).Claims(claims).Serialize()
	return res
}

func issueTestJWTWithActor(actorSubject string) string {
	// "act", Actor, ref:  [RFC8693, Section 4.1]
	// See: https://tools.ietf.org/html/rfc8693
	claims := Claims{
		"iss": "test",
		"act": map[string]string{
			"sub": actorSubject,
		},
	}
	res, _ := jwt.Signed(testSigner).Claims(claims).Serialize()
	return res
}
