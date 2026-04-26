package auth

type contextKey string

const UserIDKey contextKey = "userID"

// ClaimsKey — кладём весь *authuc.Claims из middleware. Phase 15: позволяет
// читать PlanID без users.GetByID в hot-path analytics. Старые JWT (выпущенные
// до Phase 15) имеют пустой PlanID — caller должен делать fallback на DB.
const ClaimsKey contextKey = "claims"

const BearerScheme = "bearer"
