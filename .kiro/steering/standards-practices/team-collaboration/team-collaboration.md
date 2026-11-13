# Best Practices - Spec-driven Development in Kiro

The following practices have been proven effective across various project sizes and team structures, helping organizations maintain consistency and quality in their development processes.

# Managing shared specs within a team
Teams can share artifacts (steering, hooks, and specs) by leveraging Git submodules or package references. Here are some options to consider:

## Create a central specs repository
Establish a dedicated repository for shared artifacts that multiple projects can reference

## Use Git submodules or package references
Link your central artifacts to individual projects using Git submodules, package references, or symbolic links depending on your development environment

## Implement cross-repository workflows
Develop processes for proposing, reviewing, and updating shared artifacts that affect multiple projects

# Handling multiple specs
Create separate specs for different features within your project rather than maintaining a single spec for your entire codebase.

For example, in an e-commerce application, organize your specs like this:
Sample for multi specs usage in the same project

.kiro/specs/
├── user-authentication/    # Login, signup, password reset
├── product-catalog/        # Product listing, search, filtering
├── shopping-cart/          # Add to cart, quantity updates, checkout
├── payment-processing/     # Payment gateway integration, order confirmation
└── admin-dashboard/        # Product management, user analytics

This approach helps you to:
• Work on features independently without conflicts
• Maintain focused, manageable spec documents
• Iterate on specific functionality without affecting other areas
• Collaborate with team members on different features simultaneously

# Importing existing requirements
When importing requirements or designs from another system (such as JIRA, Confluence, or Word documents), you have two options:

# Using MCP integration
Connect directly through an MCP server that supports STDIO to import requirements into your spec session

# Manual import
Copy existing requirements (e.g., foo-prfaq.md) into a new file in your repository, open a spec chat session, and use the command "#foo-prfaq.md Generate a spec from it"

# Task execution

## Perform task-wise execution
While you can execute all tasks in your tasks.md file at once, it is recommended executing tasks individually for better control and results

## Managing task execution in a collaborative environment
Option 1: Update tasks via tasks.md
Open your tasks.md file
Choose "Update tasks"
Kiro will automatically mark completed tasks

Option 2: Use Kiro's scanning capability
In a spec session, ask Kiro: "Check which tasks are already complete"
Kiro will analyze your codebase and identify implemented functionality
Tasks will be automatically marked as completed