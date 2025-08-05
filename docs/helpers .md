Voici une **documentation interne** type, pour expliquer à tes collègues l’usage de vos helpers :

# Utilisation de `utils.APISuccess` et `c.Error(apierrors...)`

## 1. Affichage d’un succès API : `utils.APISuccess`

### À quoi ça sert ?

**`utils.APISuccess`** uniformise le format des réponses HTTP 2xx, pour que toutes vos routes répondent avec le même style de payload JSON dès que l’opération a réussi.
Cela simplifie le code côté handler, la documentation Swagger, et la vie des consommateurs d’API.

### Signature

```go
func APISuccess(c *gin.Context, data interface{})
```

- **`c`** : le contexte Gin de la requête en cours.
- **`data`** : un objet ou une map contenant le contenu à renvoyer (ex : infos de l’entité créée, liste de ressources…).


### Exemple d’utilisation

**Pour retourner un succès (par exemple création d’un namespace) :**

```go
utils.APISuccess(c, gin.H{
    "message":     "Namespace created successfully",
    "name":        req.Name,
    "customer_id": customerID,
    "created_by":  email,
})
```

**Pour retourner une liste :**

```go
utils.APISuccess(c, gin.H{
    "namespaces": list,
    "count":      len(list),
})
```

**Bonnes pratiques :**

- Utiliser toujours le même style de clé ("message", "id", etc.).
- Adapter le contenu selon l’action, pas besoin de surcharge.
- Ne jamais écrire directement : `c.JSON(200, ...)` dans vos handlers métiers, **toujours passer par `APISuccess`**.


## 2. Gestion des erreurs / retours d’API : `c.Error(apierrors.XYZ("message"))`

### À quoi ça sert ?

Le couple `c.Error(apierrors....)` permet :

- D’afficher une erreur métier ou technique à l’appelant,
- De stocker au passage un log d’erreur,
- De centraliser la gestion du code HTTP (400, 401, 404, 500…).

**`apierrors`** centralise vos erreurs d’API (Gin) dans un style commun.

### Exemple d’utilisation

- **Erreur 400 (input invalide) :**

```go
c.Error(apierrors.NewBadRequest("Namespace name is required"))
// Retournera {"error": "...", "request_id": ...} avec un code 400
```

- **Erreur 401/403 (accès interdit) :**

```go
c.Error(apierrors.NewUnauthorized("Access denied"))
```

- **Erreur 404 (non trouvé) :**

```go
c.Error(apierrors.NewNotFound("Namespace not found in your tenant"))
```

- **Erreur 409 (conflit) :**

```go
c.Error(apierrors.NewConflict("The namespace name is not available."))
```

- **Erreur 500 (erreur interne) :**

```go
c.Error(apierrors.NewInternalError("Failed to query database"))
```


**Dans tous les cas** :

- Ne jamais faire un `c.AbortWithStatus(...).JSON()` directement.
- Toujours passer par vos helpers centrés (`apierrors.*`) qui gèrent le code, le message et l’audit éventuel.


## 3. Bonnes pratiques \& règles d’équipe

- **Succès →** toujours par `utils.APISuccess` (plus clair pour tous).
- **Erreur →** toujours par `c.Error(apierrors.XXX(...))` avec le message approprié.
- **Ne mélangez pas les deux (ex : un `APISuccess` post-erreur),** chaque branche du handler choisit l’un ou l’autre.


## 4. Résumé

- **APISuccess** pour toute réponse 2xx et payload constructeur.
- **apierrors.NewBadRequest/NewConflict/NewInternalError...** pour toutes les erreurs métier ou technique.
- **Toujours structurer le JSON retour de façon homogène** dans toute l’API, via ces helpers–jamais du code brut mélangeant les `c.JSON`/`c.Error` dans vos handlers métiers.


## 5. Exemple complet dans un handler

```go
if someError != nil {
    c.Error(apierrors.NewInternalError("Une erreur est survenue"))
    return
}
// Succès
utils.APISuccess(c, gin.H{
    "message": "Opération réussie",
    "id":      entity.ID,
})
```

**Choisissez le helper adapté à chaque branche selon la finalité du code.
N’hésitez pas à enrichir les messages pour l’audit ou la clarté métier, c’est la clé !**

