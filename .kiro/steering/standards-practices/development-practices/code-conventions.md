---
inclusion: always
---

# Coding Standards
This document outlines our team's coding standards and best practices.

## Naming Conventions
- Use camelCase for variables and functions
- Use PascalCase for classes and components
- Use UPPER_SNAKE_CASE for constants
- Prefix private methods with underscore (_)

## Error Handling
Always use try/catch blocks for async operations and provide meaningful error messages.
``` code-block:: typescript
 try {
  const data = await fetchData();
  return processData(data);
 } catch (error) {
  logger.error('Failed to fetch data', { error });
  throw new AppError('Unable to retrieve data. Please try again later.');
 }
```

## Code Organization
Group related functions together and order methods by:
1. Public methods
2. Protected methods
3. Private methods