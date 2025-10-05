# Update All Swagger Error Responses

## What to do:
Use your IDE's "Find and Replace in Files" feature to replace all error response types.

## Replacements needed:

### 1. Replace all error responses:
**Find:** `@@Failure 400 {object} models.ErrorRespons`
**Replace with:** `@Failure 400 {object} models.ErrorResponse`

**Find:** `@Failure 401 {object} models.ErrorResponse`
**Replace with:** `@Failure 401 {object} models.ErrorResponse`

**Find:** `@Failure 404 {object} models.ErrorResponse`
**Replace with:** `@Failure 404 {object} models.ErrorResponse`

**Find:** `@Failure 500 {object} models.ErrorResponse`
**Replace with:** `@Failure 500 {object} models.ErrorResponse`

### 2. Replace success message responses:
**Find:** `@Success 200 {object} map[string]string`
**Replace with:** `@Success 200 {object} models.MessageResponse`

**Find:** `@Success 201 {object} map[string]string`
**Replace with:** `@Success 201 {object} models.MessageResponse`

### 3. After all replacements, run:
```bash
swag init
```

### 4. Fix the generated docs.go:
Remove the `LeftDelim` and `RightDelim` fields from the SwaggerInfo struct at the end of `docs/docs.go`

## Files to update (in internal/handlers/):
- All .go files with Swagger annotations

## Result:
All error responses will show proper JSON structure like:
```json
{
  "error": "invalid_credentials"
}
```

Instead of the generic:
```json
{
  "additionalProp1": "string",
  "additionalProp2": "string",
  "additionalProp3": "string"
}
```
