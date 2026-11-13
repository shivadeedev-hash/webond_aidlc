---
inclusion: fileMatch
fileMatchPattern: ["*.js", "*.jsx", "*.ts", "*.tsx"]
---

# React Development Guidelines

## Component Structure
- Use functional components with hooks
- Separate business logic from UI components
- Follow the container/presentational pattern

## State Management
- Use React Context for global state
- Prefer useState for local component state
- Use useReducer for complex state logic

## Performance Optimization
- Memoize expensive calculations with useMemo
- Prevent unnecessary re-renders with React.memo
- Use useCallback for event handlers passed to child components
