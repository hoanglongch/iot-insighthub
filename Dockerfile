services:
  secure-api:
    image: your-docker-repo/secure-api:latest
    environment:
      - JWT_SECRET=YourSuperSecureJWTSecretValue
