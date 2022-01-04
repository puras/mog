package ctrl

/**
* @project momo-backend
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-10-03 16:24
 */

//func GetCurrentUser(c *gin.Context) (*model.User, error) {
//	user, ok := c.Get(constant.IDENTITY_KEY)
//	if !ok {
//		return &model.User{}, errdef.New(errdef.DataNotFound)
//	}
//	return user.(*model.User), nil
//}
//func GetCurrentAccount(c *gin.Context) (string, error) {
//	user, err := GetCurrentUser(c)
//	if err != nil {
//		return "", err
//	}
//	return user.Account, nil
//}
//
//func GetCurrentUserId(c *gin.Context) (string, error) {
//	user, err := GetCurrentUser(c)
//	if err != nil {
//		return "", err
//	}
//	return user.ID, nil
//}
