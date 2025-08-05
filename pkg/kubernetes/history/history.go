package history

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	kubernetesdb "github.com/Gskill75/api2/pkg/db/sqlc/kubernetes"
)

func LogNamespaceHistory(
	ctx context.Context,
	queries *kubernetesdb.Queries,
	customerID, actionType, status, namespaceName, username, createdBy, details, errorMessage string,
) {

	actionTypeEnum := kubernetesdb.KubernetesActionTypeEnum(actionType)
	statusEnum := kubernetesdb.KubernetesStatusEnum(status)

	pgErrorMessage := pgtype.Text{String: strings.TrimSpace(errorMessage), Valid: errorMessage != ""}
	fmt.Printf(
		"[HISTLOG] Will insert history: customerID=%q, actionType=%q, status=%q, nsName=%q, username=%q, createdBy=%q, details=%q\n",
		customerID, actionType, status, namespaceName, username, createdBy, errorMessage,
	)
	row, err := queries.CreateHistory(ctx, kubernetesdb.CreateHistoryParams{
		CustomerID:    customerID,
		ActionType:    actionTypeEnum,
		Status:        statusEnum,
		NamespaceName: namespaceName,
		Username:      username,
		ErrorMessage:  pgErrorMessage,
		CreatedBy:     createdBy,
		Details:       pgtype.Text{String: strings.TrimSpace(details), Valid: details != ""},
	})
	if err != nil {
		fmt.Printf("[HISTLOG] Error inserting history: %v\n", err)
	} else {
		fmt.Printf("[HISTLOG] Inserted history row: %+v\n", row)
	}

}
