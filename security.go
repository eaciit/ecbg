package ecbg

func (c *Controller) CheckAuth(requireLogin, requireAccessId, authRedirect string) {
	var (
		authErr error
	)

	if authErr != nil {
		c.Ctx.Redirect(301, authRedirect)
	}
}
