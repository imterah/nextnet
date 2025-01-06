package backendstructs

type BackendCreationRequest struct {
	Token             string      `validate:"required"`
	Name              string      `validate:"required"`
	Description       *string     `json:"description"`
	Backend           string      `validate:"required"`
	BackendParameters interface{} `json:"connectionDetails" validate:"required"`
}

type BackendLookupRequest struct {
	Token       string  `validate:"required"`
	BackendID   *uint   `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Backend     *string `json:"backend"`
}

type BackendRemovalRequest struct {
	Token     string `validate:"required"`
	BackendID uint   `json:"id" validate:"required"`
}

type ConnectionsRequest struct {
	Token string `validate:"required" json:"token"`
	Id    uint   `validate:"required" json:"id"`
}

type ProxyCreationRequest struct {
	Token           string  `validate:"required" json:"token"`
	Name            string  `validate:"required" json:"name"`
	Description     *string `json:"description"`
	Protocol        string  `validate:"required" json:"protocol"`
	SourceIP        string  `validate:"required" json:"sourceIP"`
	SourcePort      uint16  `validate:"required" json:"sourcePort"`
	DestinationPort uint16  `validate:"required" json:"destinationPort"`
	ProviderID      uint    `validate:"required" json:"providerID"`
	AutoStart       *bool   `json:"autoStart"`
}

type ProxyLookupRequest struct {
	Token           string  `validate:"required" json:"token"`
	Id              *uint   `json:"id"`
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	Protocol        *string `json:"protocol"`
	SourceIP        *string `json:"sourceIP"`
	SourcePort      *uint16 `json:"sourcePort"`
	DestinationPort *uint16 `json:"destPort"`
	ProviderID      *uint   `json:"providerID"`
	AutoStart       *bool   `json:"autoStart"`
}

type ProxyRemovalRequest struct {
	Token string `validate:"required" json:"token"`
	ID    uint   `validate:"required" json:"id"`
}

type ProxyStartRequest struct {
	Token string `validate:"required" json:"token"`
	ID    uint   `validate:"required" json:"id"`
}

type ProxyStopRequest struct {
	Token string `validate:"required" json:"token"`
	ID    uint   `validate:"required" json:"id"`
}

type UserCreationRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
	Username string `json:"username" validate:"required"`

	ExistingUserToken string `json:"token"`
	IsBot             bool   `json:"isBot"`
}

type UserLoginRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`

	Password string `json:"password" validate:"required"`
}

type UserLookupRequest struct {
	Token    string  `validate:"required"`
	UID      *uint   `json:"id"`
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Username *string `json:"username"`
	IsBot    *bool   `json:"isServiceAccount"`
}

type UserRefreshRequest struct {
	Token string `json:"token" validate:"required"`
}

type UserRemovalRequest struct {
	Token string `json:"token" validate:"required"`
	UID   *uint  `json:"uid"`
}
