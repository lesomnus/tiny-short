package bybit

import (
	"context"
	"time"
)

type UserApi interface {
	QueryApi(ctx context.Context, req UserQueryApiReq) (UserQueryApiRes, error)
	QuerySubMembers(ctx context.Context, req UserQuerySubMembersReq) (UserQuerySubMembersRes, error)
	CreateSubApiKey(ctx context.Context, req UserCreateSubApiKeyReq) (UserCreateSubApiKeyRes, error)
	// SubApiKeys(ctx context.Context, req UserSubApiKeysReq) (UserSubApiKeysRes, error)
}

type UserQueryApiReq struct{}
type UserQueryApiRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		ReadOnly    int            `json:"readOnly"`
		Permissions ApiPermissions `json:"permissions"`

		ExpiredAt time.Time `json:"expiredAt"`
		CreatedAt time.Time `json:"createdAt"`

		UserId   UserId `json:"userID"`
		IsMaster bool   `json:"isMaster"`
	} `json:"result"`
}

type UserQuerySubMembersReq struct{}
type UserQuerySubMembersRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		SubMembers []struct {
			UserId      UserId `json:"uid"`
			Username    string `json:"username"`
			MemberType  int    `json:"memberType"`
			Status      int    `json:"status"`
			AccountMode int    `json:"accountMode"`
			Remark      string `json:"remark"`
		} `json:"subMembers"`
	} `json:"result"`
}

type UserCreateSubApiKeyReq struct {
	SubUserId   UserId         `json:"subuid"`
	Note        string         `json:"note"`
	ReadOnly    int            `json:"readOnly"`
	Permissions ApiPermissions `json:"permissions"`
}
type UserCreateSubApiKeyRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		ApiKey string `json:"apiKey"`
		Secret string `json:"secret"`
	} `json:"result"`
}

// type UserSubApiKeysReq struct {
// 	SubUserId UserId `json:"subMemberId"`
// 	Limit     uint   `json:"limit"`  // Limit for data size per page. [1, 20]. Default: 20
// 	Cursor    string `json:"cursor"` // Use the nextPageCursor token from the response to retrieve the next page of the result set.
// }
// type UserSubApiKeysRes struct {
// 	ResponseBase `json:",inline"`

// 	Result struct {
// 		ApiKey string `json:"apiKey"`
// 		Secret string `json:"secret"`
// 	} `json:"result"`
// }

type userApi struct {
	client *client
}

func (a *userApi) QueryApi(ctx context.Context, req UserQueryApiReq) (res UserQueryApiRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/user/query-api")
	err = a.client.get(ctx, url, &req, &res)
	return
}

func (a *userApi) QuerySubMembers(ctx context.Context, req UserQuerySubMembersReq) (res UserQuerySubMembersRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/user/query-sub-members")
	err = a.client.get(ctx, url, &req, &res)
	return
}

func (a *userApi) CreateSubApiKey(ctx context.Context, req UserCreateSubApiKeyReq) (res UserCreateSubApiKeyRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/user/create-sub-api")
	err = a.client.post(ctx, url, &req, &res)
	return
}
