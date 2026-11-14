# Backend_YesCloth_BEFET_Final
# 👕 Yes, Cloth — Backend API

A high-performance, robust RESTful API backend for the **Yes, Cloth** e-commerce fashion platform. Built with **Go (Golang)** and **PostgreSQL**, following a clean, layered architecture with JWT authentication, role-based access control, and transactional checkout management.

---

## 🚀 Tech Stack

- **Language:** Go (Golang 1.22+)
- **Database:** PostgreSQL (with Connection Pooling via `pgx/v5`)
- **Authentication:** JWT (JSON Web Tokens) with `golang-jwt/jwt/v5`
- **Security:** Bcrypt password hashing (`golang.org/x/crypto/bcrypt`)
- **Containerization:** Docker & Docker Compose
- **Deployment:** Render

---

## 🏛️ Architecture & Folder Structure

The project follows a standard **Layered / Clean Architecture**:


---

## ✨ Features

- **Authentication & Security:**
  - Secure registration and login with bcrypt hashing.
  - Stateless JWT token-based authentication.
  - Role-Based Access Control (RBAC) separating `customer` and `admin`.
  - Self-service password reset.
- **Product & Category Management:**
  - Catalog browsing with category filtering, keyword search, and price range filters.
  - Product variant support (sizes, colors, and stock quantities).
  - Multiple image attachments per product.
- **Customer Wishlist (Favorites):**
  - Add, remove, and list favorite products per user.
- **Shopping Cart & Checkout:**
  - Persistent shopping cart per user.
  - ACID-compliant transactional checkout (creates order, records order items, snapshots shipping address, and clears cart in a single transaction).
  - User order history tracking.
- **Product Reviews & Ratings:**
  - Authenticated user reviews with 1-5 star ratings.
  - Average rating and review count aggregation.
- **Admin Management Dashboard:**
  - Full CRUD operations for Products, Categories, Users, and Reviews.
  - Real-time store order monitoring and order status management.

---

## 📡 API Endpoints

### 🔐 Authentication (`/api`)
| Method | Endpoint | Access | Description |
|---|---|---|---|
| `POST` | `/api/register` | Public | Register a new user |
| `POST` | `/api/login` | Public | Login and receive a JWT token |
| `POST` | `/api/reset-password` | Public | Reset account password |

### 🛍️ Products & Categories
| Method | Endpoint | Access | Description |
|---|---|---|---|
| `GET` | `/api/products` | Public | List & filter products |
| `GET` | `/api/categories` | Public | List all categories |

### ⭐ Favorites (Wishlist)
| Method | Endpoint | Access | Description |
|---|---|---|---|
| `GET` | `/api/favorites` | Customer | Get customer's saved wishlist |
| `POST` | `/api/favorites` | Customer | Add a product to wishlist |
| `DELETE` | `/api/favorites/{productId}` | Customer | Remove a product from wishlist |

### 🛒 Cart & Checkout
| Method | Endpoint | Access | Description |
|---|---|---|---|
| `GET` | `/api/cart` | Customer | Get customer's current cart items |
| `POST` | `/api/cart` | Customer | Add item variant to cart |
| `PUT` | `/api/cart` | Customer | Update item variant quantity in cart |
| `POST` | `/api/checkout` | Customer | Place order & checkout cart items |
| `GET` | `/api/orders/mine` | Customer | View customer order history |

### 💬 Reviews
| Method | Endpoint | Access | Description |
|---|---|---|---|
| `GET` | `/api/reviews` | Public | Get all product reviews |
| `POST` | `/api/reviews` | Customer | Submit a product review & rating |

### 👑 Admin Management
| Method | Endpoint | Access | Description |
|---|---|---|---|
| `POST` | `/api/admin/products` | Admin | Create new product & variants |
| `PUT` | `/api/admin/products/{id}` | Admin | Update product details |
| `DELETE` | `/api/admin/products/{id}` | Admin | Delete a product |
| `POST` | `/api/admin/categories` | Admin | Create a category |
| `PUT` | `/api/admin/categories/{id}` | Admin | Update a category |
| `DELETE` | `/api/admin/categories/{id}` | Admin | Delete a category |
| `GET` | `/api/admin/users` | Admin | List all registered users |
| `POST` | `/api/admin/users` | Admin | Create a new user / admin |
| `PUT` | `/api/admin/users/{id}` | Admin | Update user details or role |
| `DELETE` | `/api/admin/users/{id}` | Admin | Remove a user |
| `GET` | `/api/admin/orders` | Admin | View all customer orders |
| `PUT` | `/api/admin/reviews/{id}` | Admin | Update/moderate a review |
| `DELETE` | `/api/admin/reviews/{id}` | Admin | Delete an inappropriate review |


