# API GO

API de backend desarrollada en Go utilizando Fiber (v3), GORM y PostgreSQL con arquitectura limpia/hexagonal.

## Dependencias

- **Go**: 1.25 o superior
- **Fiber**: v3.4.0
- **Database**: PostgreSQL 16
- **ORM**: GORM
- **Contenedores**: Docker / Docker Compose

---

## Estructura

```bash
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── auth/
│   ├── customers/
│   ├── inventory/
│   ├── products/
│   ├── purchases/
│   ├── settings/
│   ├── suppliers/
│   └── user/
├── Dockerfile
├── Dockerfile.db
├── docker-compose.yml
└── init.sql
```

---

## Guía de Inicio por Primera Vez

Sigue estos pasos para poner en marcha la base de datos, ejecutar las migraciones automáticas, crear los datos iniciales (seed) y levantar la API.

### 1. Iniciar la Base de Datos PostgreSQL (con Persistencia)

**Opción A: Usando Docker Compose (Recomendado)**
```bash
docker compose up -d
```

**Opción B: Usando Docker directamente**
```bash
# Construir la imagen de la base de datos
docker build -t crm-postgres .

# Crear un volumen para la persistencia de datos
docker volume create postgres_data

# Correr el contenedor
docker run -d \
  --name crm_api_postgres \
  -p 5432:5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=api_go \
  -v postgres_data:/var/lib/postgresql/data \
  crm-postgres
```

---

### 2. Variables de Entorno

Puedes configurar las siguientes variables de entorno en tu sistema o ejecutarlas antes de iniciar el servidor (o usar los valores predeterminados):

```bash
# Servidor de Base de Datos PostgreSQL
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/api_go?sslmode=disable"

# Clave secreta para generación y validación de tokens JWT
export JWT_SECRET="supersecretkey"
```

---

### 3. Migraciones de la Base de Datos

Las migraciones de las tablas de la base de datos (`users`, `roles`, `permissions`, `settings`, `suppliers`, `products`, `purchases`, `customers`, `inventory_batches`, `inventory_logs`, `stock_adjustments`) son ejecutadas de manera **automática** por GORM (`AutoMigrate`) en el momento en que se inicia la aplicación Go.

No requieres ejecutar comandos manuales de migración SQL.

---

### 4. Iniciar la Aplicación

Para descargar las dependencias e iniciar el servidor Go:

```bash
# Descargar dependencias del proyecto
go mod download

# Iniciar la API
go run cmd/api/main.go
```

La API quedará escuchando por defecto en `http://localhost:3000`.

---

### 5. Semilla de Datos Iniciales (Seed / Primer Usuario Admin)

Para crear tu primer usuario Administrador por primera vez y poder autenticarte en el sistema:

#### A. Registro del Usuario Administrador Inicial
Realiza una petición HTTP `POST` a `/auth/register`:

```bash
curl -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Administrador",
    "email": "admin@example.com",
    "password": "AdminPassword123!"
  }'
```

#### B. Inicio de Sesión (Obtener JWT Token)
Realiza una petición HTTP `POST` a `/auth/login`:

```bash
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "AdminPassword123!"
  }'
```

La respuesta devolverá un token JWT que deberás incluir en el encabezado `Authorization: Bearer <token>` para los endpoints protegidos.

---

## Verificación del Estado de los Servicios

Puedes probar el correcto funcionamiento de los módulos consultando los siguientes endpoints:

- **Autenticación**: `POST /auth/login`
- **Productos**: `GET /api/v1/products`
- **Inventario**: `GET /api/v1/inventory/metrics`
- **Clientes**: `GET /api/v1/customers`
- **Compras**: `GET /api/v1/purchases`
- **Proveedores**: `GET /api/v1/suppliers`
