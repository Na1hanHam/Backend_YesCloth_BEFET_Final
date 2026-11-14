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
