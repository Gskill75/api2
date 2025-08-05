Voici un bloc de documentation (expliqué ET exemples) pour que tes collègues sachent comment correctement utiliser la fonction de journalisation d’historique `LogNamespaceHistory` dans vos handlers.
Ce modèle est conçu pour être inséré dans votre `/pkg/utils/history.go` (ou autre README de l’équipe).

# Utilisation de `LogNamespaceHistory`

## À quoi ça sert ?

La fonction `LogNamespaceHistory` permet de tracer toutes les actions importantes sur les namespaces Kubernetes (création, suppression, échecs, etc.) dans la table `kubernetes_history` pour audit, debug ou supervision.
**Toute action critiquable métier (succès ou erreur) doit être tracée, pour avoir un historique fiable dans la BDD.**

## Signature de la fonction

```go
func LogNamespaceHistory(
    ctx context.Context,
    queries *kubernetesdb.Queries,
    customerID, actionType, status, namespaceName, username, createdBy, details, errorMessage string,
)
```


### Paramètres

| Paramètre | Type | Description |
| :-- | :-- | :-- |
| `ctx` | Context | Le contexte d’exécution (ex : `c.Request.Context()`) |
| `queries` | *Queries | L’instance SQLC pour la base (ne jamais passer nil) |
| `customerID` | string | L’identifiant client concerné |
| `actionType` | string | Le type d’action (`create`, `delete`, `admin_delete`, etc.) |
| `status` | string | `"success"` ou `"error"` |
| `namespaceName` | string | Nom du namespace impacté (ou `""` si pas disponible) |
| `username` | string | Opérateur effectuant l’action (généralement email) |
| `createdBy` | string | Même valeur que username, ou propriétaire logique |
| `details` | string | Description métier/contextuelle de l’action (custom, facultatif) |
| `errorMessage` | string | Erreur technique détaillée, ex : `err.Error()` (facultatif) |

## Exemple d’utilisation

### Pour un succès

```go
history.LogNamespaceHistory(
    c.Request.Context(),
    queries,
    customerID,
    "create",      // actionType
    "success",     // status
    req.Name,      // namespaceName
    email,         // username (ex: issu du JWT)
    email,         // createdBy
    "",            // details
    "",            // errorMessage
)
```


### Pour un échec métier/technique

```go
if err != nil {
    utils.LogNamespaceHistory(
        c.Request.Context(),
        queries,
        customerID,
        "delete",       // actionType
        "error",        // status
        name,           // namespaceName
        email,
        email,
        "Erreur technique lors de la suppression", // details
        err.Error(),    // errorMessage (texte brut de l'erreur Go)
    )
    // ... gérer l'erreur API
}
```


### Pour un input invalide (ex: nom de namespace manquant)

```go
if name == "" {
    utils.LogNamespaceHistory(
        c.Request.Context(),
        queries,
        customerID,
        "delete",
        "error",
        "",           // pas de nom dispo
        email,
        email,
        "Namespace name is empty",
        "Missing namespace name",
    )
    // ... gérer l'erreur API
}
```


## Règles d’utilisation

- Appeler la fonction **juste avant chaque `return`** où une erreur ou un succès doit être tracé (succès en fin de handler, échec à chaque branche d’erreur).
- **Toujours renseigner 10 paramètres**, détaillez si possible le contexte (et éviter les nil dereference !).
- Pour chaque handler, tracer :
    - Les succès importants
    - Tous les cas d’erreur métier/technique
- Les valeurs de `actionType` et `status` doivent correspondre à vos ENUMS SQL (ex: "create", "delete", "success", "error", "admin_delete"…).


## Bonnes pratiques

- Si vous n’avez pas de valeur pour un champ (ex : nom du namespace sur un mauvais JSON), passez `""` ou `"unknown"` pour garder une cohérence d’audit.
- Ne jamais appeler `.Error()` sur une variable err si vous n’êtes pas certain qu’elle n'est pas nil.
- Vérifiez que la colonne existe bien dans la base (alignement schema/requête SQL).


## À intégrer dans un PR/code review

- **Toute nouvelle action sur un namespace doit être historiée.**
- Ajoutez le log d’historique **dans chaque handler concerné**.
- Si vous ajoutez un nouveau champ d’audit, pensez à modifier la table SQL, la requête SQLC, et cette doc.

**En résumé :**
“**LogNamespaceHistory** doit être appelée sur chaque action majeure sur les namespaces pour conserver un audit métier/technique exhaustif dans la base.”


