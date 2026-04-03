# Multi-stage build for optimal image size and security

# Stage 1: Dependencies
FROM oven/bun:1.3.11-alpine AS deps
WORKDIR /app

# Copy package files
COPY package.json bun.lock ./

# Install production dependencies only
RUN bun install --frozen-lockfile --production

# Stage 2: Builder
FROM oven/bun:1.3.11-alpine AS builder
WORKDIR /app

# Copy package files
COPY package.json bun.lock ./

# Install all dependencies (including devDependencies for build)
RUN bun install --frozen-lockfile

# Copy source code
COPY . .

# Build the application with BUILD_TARGET=node to use the node adapter
RUN bun run build:node

# Stage 3: Runner (Production)
FROM oven/bun:1.3.11-alpine AS runner
WORKDIR /app

# Install dumb-init for proper signal handling
RUN apk add --no-cache dumb-init

# Create a non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S observer -u 1001 -G nodejs

# Copy necessary files from builder
COPY --from=builder --chown=observer:nodejs /app/build ./build
COPY --from=builder --chown=observer:nodejs /app/package.json ./

# Copy production dependencies from deps stage
COPY --from=deps --chown=observer:nodejs /app/node_modules ./node_modules

# Switch to non-root user
USER observer

# Expose the application port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD bun -e "fetch('http://127.0.0.1:' + (process.env.PORT || '3000')).then(r => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))"

# Use dumb-init to handle signals properly
ENTRYPOINT ["dumb-init", "--"]

# Start the application
CMD ["bun", "run", "start"]
