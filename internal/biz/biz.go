/*
 * @Descripttion:
 * @version:
 * @Date: 2023-04-29 22:30:30
 * @LastEditTime: 2023-07-02 00:48:31
 */
package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewGpt, NewRoleTemplateUsecase, NewMeetingTemplateUsecase, NewImageUsecase, NewMeetingUsecase)
