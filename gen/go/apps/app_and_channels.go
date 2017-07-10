/* 
 * Vendor API V1
 *
 * Apps documentation
 *
 * OpenAPI spec version: 1.0.0
 * 
 * Generated by: https://github.com/swagger-api/swagger-codegen.git
 */

package swagger

type AppAndChannels struct {

	App App `json:"App"`

	// Channels of the app
	Channels []AppChannel `json:"Channels"`
}