# 🛰️ API Self-Service Cloud – README Technique

## Présentation

Cette API écrite en **Go** fournit un portail unifié pour la gestion self-service des ressources cloud de SIGMA :

- **Kubernetes** : gestion complète des namespaces (création, suppression, audit, RBAC…), avec historique fiabilisé en base SQL.
    
- **DBaaS PostgreSQL** : provisionnement as-a-service (via AWX), gestion autonome des bases avec suivi d’état et audit intégré.
    
- **Harbor/Registry** : opérations sur les registries d’images (création, suppression, audit...).
    
- **Extensible** : architecture modulaire et versionnée, prête à accueillir de nouveaux domaines sans régression.
    

### ✨ Points techniques clés

- **Go + Gin (REST)** : API rapide et modulaire avec des handlers minces.
    
- **Clean Architecture** : séparation stricte des couches (Solution → Service → Handler).
    
- **Inversion de dépendances (DI)** : chaque couche reste testable indépendamment.
    
- **SQLC (pgx/v5)** : code typé et maintenable généré à partir de schémas/queries SQL.
    
- **Goose** : migrations SQL versionnées, transactionnelles et déployables en CI/CD.
    
- **Swagger** : documentation OpenAPI générée automatiquement via annotations.
    
- **Healthchecks** : endpoints `/healthz` intégrés pour PostgreSQL, Kubernetes, Harbor...
    
- **Audit centralisé** : traçabilité métier via base SQL + logs structurés (`klog/v2`).
    

---

## 🎯 Objectifs

- Fournir une **base scalable, testable et maintenable** pour l’expansion des services cloud.
    
- **Renforcer la sécurité et l’auditabilité** (JWT, RBAC, logs, historique).
    
- Assurer la **lisibilité, la cohérence et la factorisation** du code (clean layers, DI, Swagger...).
    

---

## 📁 Arborescence projet (extrait)

``` bash
pkg/
├── kubernetes/               # Module Kubernetes v1
│   ├── kubernetes.go         # Entrypoint (Solution) : DI/routes
│   ├── client/               # Instanciation client K8s
│   ├── handler/              # REST controllers
│   ├── service/              # Logique métier (DB, K8s, audit)
│   ├── history/              # Gestion d’historique/audit
│   └── ...
├── kubernetes_v2/            # Version majeure suivante (v2)
├── dbaas/                    # Module DBaaS (PostgreSQL)
├── harbor/                   # Module Harbor
├── db/                       # Requêtes SQL + modèles générés
│   ├── dbaas/, harbor/, ...
│   └── sqlc/                 # Code Go généré (via sqlc)
├── docs/
├── errors/
├── health/
└── ...
main.go                       # Initialisation, routing multi-version

```


---

## 🧱 Architecture & Patterns

### Clean layering

|Couche|Rôle|
|---|---|
|**Solution**|Initialisation du module, wiring des dépendances, routing versionné|
|**Service**|Logique métier : accès DB, appels K8s, validations, audit|
|**Handler**|Contrôleur REST simple : parse les requêtes, appelle le service|
|**Client**|Création de clients externes (K8s, AWX...), injectés dans les services|
|**History**|Centralisation des audits : appel systématique depuis les services/handlers|

### Versioning natif

- Chaque nouvelle version d’API = nouveau dossier (`kubernetes_v2/`, `dbaas_v2/`, etc.)
    
- Aucun écrasement du code précédent.
    
- Le `main.go` gère automatiquement les routes versionnées :  
    `/api/<solution>/<version>/...`
    

### Génération SQLC

- `sqlc generate` produit un package Go par domaine (`kubernetesdb`, `harbordb`…)
    
- Séparation des `schema/` et `query/`
    
- Typage fort, testé, lisible et maintenable
    

---

## 🛠️ Bonnes pratiques de développement

### Ajout/modification d’une fonctionnalité

1. Ajouter la logique métier dans `/service/`
    
2. Créer le handler REST dans `/handler/`
    
3. Ajouter wiring et DI dans le fichier `*.go` du module (Solution)
    
4. S'assurer que le routing est versionné dans `main.go`
    
5. Intégrer un appel à l’audit via `History`
    
6. Documenter chaque endpoint avec un bloc Swagger
    
7. Logger via `klog/v2` (niveau en fonction de la criticité)
    

### Exemple d’annotation Swagger

``` go
// HelloHandler godoc
// @Summary      Hello world message
// @Description  Retourne un message de test pour l'API v2
// @Tags         kubernetes-v2
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /kubernetes/v2/hello [get]
// @Security     Bearer

```
---

## 🚀 Démarrage local

1. - Copier `config.dev.yaml` → `config.yaml` puis adapter si besoin.
    
2. Démarrer :
    
``` bash
go run main.go --config config.yaml
# ou build
make build
```


3. Accès à la documentation Swagger : [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
    
4. Healthchecks : `GET /healthz`
    
5. API versionnée : `/api/kubernetes/v1/...`, `/api/kubernetes/v2/...`, etc.
    

---

## ⚙️ Configuration SQLC (extrait)

``` yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "./pkg/db/kubernetes/schema"
    queries: "./pkg/db/kubernetes/query"
    gen:
      go:
        package: "kubernetesdb"
        out: "./pkg/db/sqlc/kubernetes"
        sql_package: "pgx/v5"
  ...
```

Après chaque modification SQL :
``` bash
sqlc generate
```

---

## ❓ FAQ & conseils

**Où coder quoi ?**

|Type de logique|Dossier|
|---|---|
|Logique métier|`/service/`|
|Contrôleurs REST|`/handler/`|
|Clients techniques|`/client/`|
|Traçabilité/audit|`/history/`|

**Comment ajouter un module/version ?**

1. Créer un nouveau dossier (ex. `kubernetes_v2/`)
    
2. Reproduire l’architecture Solution/Service/Handler
    
3. Ajouter la Solution dans `main.go`
    

**Audit/tracabilité ?**

- Appels à `LogNamespaceHistory()` dans tous les cas critiques
    
- Logging uniforme via `klog/v2`
    

**Debug et test ?**

- Centraliser les appels externes dans `/client/` pour faciliter le mock
    
- Ajouter des sous-packages métier si nécessaire (ex. `namespace/`, `rbac/`, `quota/`)
    
- Factoriser les modèles communs dans un dossier `model/` (par module)
    

---
📄 [Consulter les documentations complemantaires dans le dossier `docs/`](./docs/)
