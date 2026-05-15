# AIO Fund Client Dashboard - Design System

## Overview
This design system defines the visual language, components, and patterns for the AIO Fund Client Dashboard. It ensures consistency across all pages and components.

---

## Color Palette

### Primary Colors
```css
--primary-navy: #0B1930    /* Dark navy - sidebar, headers, primary text */
--primary-gold: #C6A661    /* Gold - accents, CTAs, active states */
```

### Background Colors
```css
--bg-main: #F9FAFB        /* Main app background */
--bg-light: #F7F8FA       /* Alternative background */
--bg-white: #FFFFFF       /* Card backgrounds */
```

### Text Colors
```css
--text-primary: #1E1E1E   /* Primary text */
--text-secondary: #5A5A5A /* Secondary text, labels */
--text-navy: #0B1930      /* Headers, important text */
```

### Status Colors
```css
--status-scheduled: #C6A661  /* Gold - upcoming/scheduled */
--status-completed: #2E7D32  /* Green - completed */
--status-cancelled: #D32F2F  /* Red - cancelled/error */
```

### Neutral Colors
```css
--border-light: rgba(0, 0, 0, 0.1)
--gray-200: #E5E7EB
--gray-100: #F3F4F6
```

---

## Typography

### Font Family
```css
font-family: Inter, system-ui, sans-serif;
```

### Font Sizes (Auto-applied via globals.css)
- **h1**: 2xl (medium weight, line-height 1.5)
- **h2**: xl (medium weight, line-height 1.5)
- **h3**: lg (medium weight, line-height 1.5)
- **h4**: base (medium weight, line-height 1.5)
- **p**: base (normal weight, line-height 1.5)
- **button**: base (medium weight, line-height 1.5)
- **input**: base (normal weight, line-height 1.5)

### Font Weights
- **Normal**: 400
- **Medium**: 500

### Important Note
**DO NOT** use Tailwind font classes (text-2xl, text-xl, font-bold, etc.) unless specifically requested. The typography is handled automatically via the globals.css file.

---

## Layout Structure

### Page Container
```tsx
<div className="p-8">
  {/* Page content */}
</div>
```
- Standard padding: `p-8` (32px)
- Background inherits from layout: `#F9FAFB`

### Page Header Pattern
```tsx
<div className="flex items-center justify-between mb-6">
  <div>
    <h2 className="text-[#1E1E1E] mb-1">Page Title</h2>
    <p className="text-[#5A5A5A]">Page description</p>
  </div>
  <Button className="bg-[#C6A661] hover:bg-[#B39551] text-white">
    Action
  </Button>
</div>
```

---

## Components

### Cards
**Standard Card Pattern:**
```tsx
<Card style={{ borderRadius: '10px', boxShadow: '0 2px 6px rgba(0,0,0,0.05)' }}>
  <CardHeader>
    <CardTitle>Title</CardTitle>
  </CardHeader>
  <CardContent>
    {/* Content */}
  </CardContent>
</Card>
```

**Specifications:**
- Border radius: `10px`
- Box shadow: `0 2px 6px rgba(0,0,0,0.05)`
- Background: `#FFFFFF`
- Border: subtle, via shadow

### Buttons

**Primary Button:**
```tsx
<Button
  className="bg-[#C6A661] hover:bg-[#B39551] text-white"
  style={{ borderRadius: '10px' }}
>
  Label
</Button>
```

**Secondary Button:**
```tsx
<Button
  variant="outline"
  style={{ borderRadius: '10px' }}
>
  Label
</Button>
```

**Icon Button:**
```tsx
<Button
  variant="outline"
  size="icon"
  style={{ borderRadius: '10px' }}
>
  <Icon className="w-4 h-4" />
</Button>
```

**Specifications:**
- Border radius: `10px`
- Primary: Gold background (#C6A661), white text
- Primary hover: Darker gold (#B39551)
- Secondary: Outlined with border
- Icon size: `w-4 h-4` or `w-5 h-5`

### Inputs

**Text Input:**
```tsx
<Input
  placeholder="Placeholder text"
  style={{ borderRadius: '10px' }}
/>
```

**Select Dropdown:**
```tsx
<Select>
  <SelectTrigger style={{ borderRadius: '10px' }}>
    <SelectValue placeholder="Select..." />
  </SelectTrigger>
  <SelectContent>
    <SelectItem value="option1">Option 1</SelectItem>
  </SelectContent>
</Select>
```

**Specifications:**
- Border radius: `10px`
- Background: `#F3F3F5` (input-background)
- Border: `rgba(0, 0, 0, 0.1)`

### Badges

**Status Badge Pattern:**
```tsx
<Badge
  variant="outline"
  style={{
    backgroundColor: '#C6A66120',  // 20% opacity
    borderColor: '#C6A661',
    color: '#C6A661',
    borderRadius: '4px',
  }}
>
  Status
</Badge>
```

**Status Colors:**
- Scheduled: Gold (#C6A661)
- Completed: Green (#2E7D32)
- Cancelled: Red (#D32F2F)

### Avatars
```tsx
<Avatar className="w-10 h-10 border-2 border-[#C6A661]">
  <AvatarFallback className="bg-[#C6A661] text-white">
    JD
  </AvatarFallback>
</Avatar>
```

### Modals/Dialogs
```tsx
<Dialog>
  <DialogContent style={{ borderRadius: '10px' }}>
    <DialogHeader>
      <DialogTitle>Title</DialogTitle>
      <DialogDescription>Description</DialogDescription>
    </DialogHeader>
    {/* Content */}
  </DialogContent>
</Dialog>
```

---

## Navigation

### Sidebar
- Width: `90px`
- Background: `#0B1930` (primary navy)
- Icon color (inactive): `white/60`
- Icon color (active): `#C6A661` (gold)
- Icon color (hover): `white`
- Logo background: `#C6A661`

### Top Header
- Height: `64px`
- Background: `#FFFFFF`
- Border: `border-b border-gray-200`
- Padding: `px-8 py-4`
- Title color: `#0B1930`

---

## Spacing System

### Standard Spacing
- Extra small: `gap-1` (4px)
- Small: `gap-2` (8px)
- Medium: `gap-4` (16px)
- Large: `gap-6` (24px)
- Extra large: `gap-8` (32px)

### Padding/Margin Scale
- Page container: `p-8`
- Card padding: `p-6` (CardContent default)
- Section spacing: `mb-6` between major sections
- Element spacing: `mb-4` between related elements

### Grid Layouts
```tsx
// 3-column layout (KPI cards)
<div className="grid grid-cols-1 md:grid-cols-3 gap-6">

// 2-column layout (Messages + Actions)
<div className="grid grid-cols-1 lg:grid-cols-2 gap-6">

// 4-column layout (Trust info)
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
```

---

## Interactive States

### Hover States
- Buttons: Slightly darker shade
- Cards: `hover:bg-[#F7F8FA]` or `hover:shadow-sm`
- Links: Underline or color change
- Icons: Opacity or color change

### Active States
- Navigation items: Gold color (`#C6A661`)
- Selected items: Background color (`#F9FAFB`)
- Focus: Ring with `--ring` color

### Disabled States
- Opacity: `opacity-50`
- Cursor: `cursor-not-allowed`

---

## Icons

### Library
**lucide-react** - https://lucide.dev

### Standard Sizes
- Small: `w-4 h-4` (16px)
- Medium: `w-5 h-5` (20px)
- Large: `w-6 h-6` (24px)

### Usage Examples
```tsx
import { Calendar, Plus, Search, ChevronRight } from 'lucide-react';

<Calendar className="w-5 h-5" />
<Plus className="w-4 h-4" />
```

---

## Patterns

### Filter Row Pattern
```tsx
<Card style={{ borderRadius: '10px', boxShadow: '0 2px 6px rgba(0,0,0,0.05)' }}>
  <CardContent className="pt-6">
    <div className="grid grid-cols-1 md:grid-cols-12 gap-4">
      {/* Filters */}
    </div>
  </CardContent>
</Card>
```

### Table Pattern
```tsx
<Table>
  <TableHeader>
    <TableRow>
      <TableHead>Column</TableHead>
    </TableRow>
  </TableHeader>
  <TableBody>
    <TableRow>
      <TableCell>Data</TableCell>
    </TableRow>
  </TableBody>
</Table>
```

### Tab Pattern
```tsx
<Tabs defaultValue="tab1" className="w-full">
  <TabsList className="mb-6">
    <TabsTrigger value="tab1">Tab 1</TabsTrigger>
    <TabsTrigger value="tab2">Tab 2</TabsTrigger>
  </TabsList>
  <TabsContent value="tab1">
    {/* Content */}
  </TabsContent>
</Tabs>
```

---

## Accessibility

### ARIA Labels
- Always include `title` attributes on icon buttons
- Use `aria-label` for non-text buttons
- Include DialogDescription in all modals

### Keyboard Navigation
- All interactive elements must be keyboard accessible
- Logical tab order
- Enter key for form submissions

### Focus States
- Visible focus indicators
- Skip to main content link (if needed)

---

## Responsive Design

### Breakpoints (Tailwind defaults)
- `sm`: 640px
- `md`: 768px
- `lg`: 1024px
- `xl`: 1280px
- `2xl`: 1536px

### Mobile-First Patterns
```tsx
// Stack on mobile, grid on desktop
className="grid grid-cols-1 md:grid-cols-3 gap-6"

// Hide on mobile, show on large screens
className="hidden lg:block"

// Full width on mobile, fixed on desktop
className="w-full lg:w-auto"
```

---

## Toast Notifications

### Error Toasts Only
The system uses **error toasts only** with helpful user guides.

```tsx
import { toast } from "sonner@2.0.3";

toast.error("Error message", {
  description: "Helpful guide on how to fix the issue"
});
```

**No success or info toasts** - keeps the UI clean and focused.

---

## Form Validation

### Input Error States
```tsx
<Input
  className="border-red-500"
  aria-invalid="true"
/>
<p className="text-sm text-red-500 mt-1">Error message</p>
```

---

## Animation

### Library
**motion/react** (formerly Framer Motion)

```tsx
import { motion } from 'motion/react';

<motion.div
  initial={{ opacity: 0 }}
  animate={{ opacity: 1 }}
  transition={{ duration: 0.3 }}
>
  Content
</motion.div>
```

### Standard Transitions
- Duration: 200-300ms
- Easing: Default (ease-in-out)
- Hover: Scale slightly or change color

---

## Data Formatting

### Dates
- Use `date-fns` for date manipulation
- Display format: `format(date, 'MMM d, yyyy')` → "Mar 29, 2025"
- Time format: `h:mm a` → "2:00 PM"

### Currency
```tsx
const formatted = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD'
}).format(amount);
// $75,000.00
```

### Numbers
```tsx
const formatted = number.toLocaleString('en-US');
// 1,234,567
```

---

## Best Practices

### Component Organization
1. Import statements
2. Type definitions
3. Data/constants
4. Component function
5. Export

### File Naming
- Components: PascalCase (`DashboardPage.tsx`)
- Utilities: camelCase (`utils.ts`)
- Constants: UPPER_SNAKE_CASE

### State Management
- Use `useState` for local component state
- Lift state up when needed by multiple components
- Pass callbacks for state updates

### Code Quality
- Prefer functional components
- Use TypeScript interfaces for props
- Extract reusable logic into custom hooks
- Keep components focused and single-purpose

---

## Component Library

### ShadCN Components Available
See `/components/ui` directory for full list:
- accordion, alert, alert-dialog
- avatar, badge, button
- calendar, card, carousel
- checkbox, dialog, drawer
- dropdown-menu, form, input
- select, table, tabs
- And many more...

### Usage
```tsx
import { Button } from './components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './components/ui/card';
```

---

## Version Control

### Current Stack
- React 18+
- Tailwind CSS 4.0
- TypeScript
- date-fns
- lucide-react
- motion/react
- ShadCN UI components

---

## Summary Checklist

When creating new components:
- ✅ Use `p-8` for page container
- ✅ Cards: `borderRadius: '10px'`, `boxShadow: '0 2px 6px rgba(0,0,0,0.05)'`
- ✅ Buttons: Gold primary (`#C6A661`), 10px radius
- ✅ Text colors: `#1E1E1E`, `#5A5A5A`, `#0B1930`
- ✅ NO font size/weight Tailwind classes (unless requested)
- ✅ Consistent spacing: `gap-4`, `gap-6`, `mb-6`
- ✅ Icons from lucide-react
- ✅ Responsive grid layouts
- ✅ Accessible with proper ARIA labels
- ✅ Error toasts only (no success/info)
