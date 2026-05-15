# 🎨 White-Label Branding Guide

## Overview

The AIO Fund Client Dashboard supports white-labeling, allowing you to customize the application branding to match your company identity while optionally displaying "Powered by AIO Fund" attribution.

## Quick Start

### 1. Locate the Configuration File

Navigate to `/config/branding.ts` in your project.

### 2. Update the Branding Configuration

Edit the `brandingConfig` object with your company details:

```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'Your Company Name',
  companyInitials: 'YCN',
  showPoweredBy: true,
};
```

### 3. Refresh the Application

After saving your changes, refresh the browser to see your new branding.

---

## Configuration Options

### Required Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `companyName` | string | Your company's full legal or business name | `"Nexxess Business Advisors"` |
| `companyInitials` | string | 2-4 letter acronym for your company | `"NBA"` |
| `showPoweredBy` | boolean | Show "Powered by AIO Fund" badge | `true` |

### Optional Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `primaryColor` | string | Override the default gold accent color (#C6A661) | `"#1E3A8A"` |
| `logoUrl` | string | Path to your custom logo image | `"/assets/logo.png"` |

---

## Example Configurations

### Example 1: Nexxess Business Advisors

```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'Nexxess Business Advisors',
  companyInitials: 'NBA',
  showPoweredBy: true,
};
```

**Result:**
- Logo displays "NBA" on gold background
- Company name appears in sidebar and login page
- "Powered by AIO Fund" appears below logo

### Example 2: Wealth Management Group

```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'Wealth Management Group',
  companyInitials: 'WMG',
  showPoweredBy: true,
  primaryColor: '#1E40AF', // Blue accent
};
```

**Result:**
- Logo displays "WMG" on blue background
- Blue accent color throughout the app
- "Powered by AIO Fund" badge shown

### Example 3: Full White-Label (No Attribution)

```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'Premier Trust Services',
  companyInitials: 'PTS',
  showPoweredBy: false,
};
```

**Result:**
- Logo displays "PTS" on gold background
- No "Powered by AIO Fund" attribution
- Appears as a fully independent application

### Example 4: Custom Logo Image

```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'Global Financial Advisors',
  companyInitials: 'GFA',
  showPoweredBy: true,
  logoUrl: '/assets/gfa-logo.png',
};
```

**Result:**
- Custom logo image displayed instead of initials
- Company name and "Powered by AIO Fund" shown
- Professional branded appearance

---

## Where Branding Appears

Your branding configuration affects the following areas:

### 1. **Sidebar Navigation**
- Company logo/initials at the top
- Company name below logo
- "Powered by AIO Fund" badge (if enabled)

### 2. **Login Page**
- Large logo/initials in header
- Company name as main title
- "Powered by AIO Fund" subtitle (if enabled)

### 3. **Guest Learning Center**
- Company name in welcome banner
- "Powered by AIO Fund" attribution (if enabled)

### 4. **All Authenticated Pages**
- Consistent branding throughout the application
- Company logo always visible in sidebar

---

## Design Guidelines

### Logo/Initials Best Practices

- **2-3 letters** works best for initials (e.g., "NBA", "AIO")
- **4 letters maximum** to maintain readability
- Use **uppercase** for professional appearance
- Avoid special characters or numbers

### Company Name Guidelines

- Use your **official business name**
- Keep it **concise** (under 30 characters preferred)
- Full legal name or "doing business as" (DBA) name both work
- Will be truncated in sidebar if too long

### Custom Logo Images

If using a custom logo (`logoUrl`):
- **Recommended size:** 64x64 pixels (minimum)
- **Format:** PNG with transparency preferred
- **Aspect ratio:** Square (1:1) works best
- **File location:** Place in `/public` or `/assets` folder
- **Path format:** Use relative paths like `/assets/logo.png`

### Color Selection

If overriding `primaryColor`:
- Use **hexadecimal format** (e.g., `#1E40AF`)
- Choose colors with good **contrast** against white and dark backgrounds
- Test with your logo to ensure cohesion
- Default gold (`#C6A661`) is optimized for trust/financial services

---

## Quick Preset Templates

Copy and paste these ready-to-use configurations:

### AIO Fund (Default)
```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'AIO Fund',
  companyInitials: 'AIO',
  showPoweredBy: false,
};
```

### Nexxess Business Advisors
```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'Nexxess Business Advisors',
  companyInitials: 'NBA',
  showPoweredBy: true,
};
```

### Custom Firm
```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'Your Firm Name Here',
  companyInitials: 'YFN',
  showPoweredBy: true,
  // primaryColor: '#YOUR_COLOR', // Uncomment to customize
  // logoUrl: '/path/to/logo.png', // Uncomment to use custom logo
};
```

---

## Viewing Current Configuration

You can view your current branding configuration in the application:

1. Log into the dashboard
2. Navigate to **Profile & Settings** (gear icon)
3. Click the **White-Label Branding** tab
4. See your current configuration and preview

---

## Troubleshooting

### Branding not updating?

1. **Clear browser cache** (Ctrl+Shift+R or Cmd+Shift+R)
2. Verify the config file is saved properly
3. Check for syntax errors in the TypeScript file
4. Ensure the application has been reloaded

### Logo not displaying?

1. Check the `logoUrl` path is correct
2. Ensure the image file exists in the specified location
3. Verify the image format is supported (PNG, JPG, SVG)
4. Check browser console for 404 errors

### Colors not changing?

1. Use hexadecimal format: `#RRGGBB`
2. Include the `#` symbol
3. Verify the color is being applied in developer tools
4. Check for CSS specificity issues

---

## Support

For assistance with white-label configuration:

1. Check the **White-Label Branding** tab in Settings
2. Review this guide for examples
3. Contact your AIO Fund technical support team
4. Submit a ticket for custom branding assistance

---

## License & Attribution

When using `showPoweredBy: true`, the "Powered by AIO Fund" badge acknowledges that this portal is built on the AIO Fund platform while maintaining your brand identity.

For full white-labeling without attribution, set `showPoweredBy: false` (requires appropriate licensing agreement).
