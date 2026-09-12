package v1

// Services is everything the API exposes, which is everything the app uses.
//
// It lives here rather than beside one of the doors because every door serves the same
// set: Connect for the browser, the gateway for REST, and MCP for an assistant. A
// second list would be a second place for a service to go missing from.
type Services struct {
	Activity *ActivityService
	Auth     *AuthService
	Instance *InstanceService
	List     *ListService
	Member   *MemberService
	Public   *PublicService
	Request  *RequestService
	Token    *TokenService
}
