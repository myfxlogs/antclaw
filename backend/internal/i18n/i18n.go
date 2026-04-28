package i18n

import "context"

func T(ctx context.Context, key string, _ map[string]interface{}) string {
return key
}
