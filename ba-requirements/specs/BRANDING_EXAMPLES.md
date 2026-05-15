# 🎨 Branding Configuration Examples

This document provides visual examples and code snippets for different white-label branding configurations.

---

## 🏢 Configuration Presets

### 1. Default AIO Fund

**Use Case:** Official AIO Fund branding with no attribution needed

```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'AIO Fund',
  companyInitials: 'AIO',
  showPoweredBy: false,
};
```

**Visual Preview:**
```
┌─────────────┐
│             │
│    ┌───┐    │
│    │AIO│    │  ← Gold background (#C6A661)
│    └───┘    │
│             │
│  AIO Fund   │  ← White text (10px)
│             │
└─────────────┘
```

---

### 2. Nexxess Business Advisors

**Use Case:** Partner firm with AIO attribution

```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'Nexxess Business Advisors',
  companyInitials: 'NBA',
  showPoweredBy: true,
};
```

**Visual Preview:**
```
┌─────────────┐
│             │
│    ┌───┐    │
│    │NBA│    │  ← Gold background
│    └───┘    │
│             │
│   Nexxess   │  ← White text
│  Business   │
│  Advisors   │
│             │
│ Powered by  │  ← Gray text (40% opacity)
│  AIO Fund   │
│             │
└─────────────┘
```

---

### 3. Custom with Blue Accent

**Use Case:** Firm with custom brand colors

```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'Wealth Management Group',
  companyInitials: 'WMG',
  showPoweredBy: true,
  primaryColor: '#1E40AF', // Blue
};
```

**Visual Preview:**
```
┌─────────────┐
│             │
│    ┌───┐    │
│    │WMG│    │  ← Blue background (#1E40AF)
│    └───┘    │
│             │
│   Wealth    │
│ Management  │
│    Group    │
│             │
│ Powered by  │
│  AIO Fund   │
│             │
└─────────────┘
```

---

### 4. Custom Logo Image

**Use Case:** Firm with professional logo

```typescript
export const brandingConfig: BrandingConfig = {
  companyName: 'Premier Trust Services',
  companyInitials: 'PTS',
  showPoweredBy: true,
  logoUrl: '/assets/premier-logo.png',
};
```

**Visual Preview:**
```
┌─────────────┐
│             │
│  ┌───────┐  │
│  │ LOGO  │  │  ← Custom PNG/SVG image
│  │ IMAGE │  │
│  └───────┘  │
│             │
│   Premier   │
│    Trust    │
│  Services   │
│             │
│ Powered by  │
│  AIO Fund   │
│             │
└─────────────┘
```

---

## 📱 Login Page Branding

Each configuration affects the login page differently:

### Standard Configuration
```
┌─────────────────────────────────────┐
│                                     │
│     ┌──────┐                        │
│     │      │                        │
│     │ NBA  │  ← 64x64px logo       │
│     │      │                        │
│     └──────┘                        │
│                                     │
│  Nexxess Business Advisors          │  ← 32px heading
│                                     │
│  Manage your trust activation       │
│  process with confidence            │
│                                     │
│  Powered by AIO Fund                │  ← If enabled
│                                     │
└─────────────────────────────────────┘
```

---

## 🎨 Color Customization

### Available Color Options

You can customize the primary accent color used throughout the application:

```typescript
primaryColor: '#C6A661'  // Gold (Default)
primaryColor: '#1E40AF'  // Blue
primaryColor: '#059669'  // Green
primaryColor: '#DC2626'  // Red
primaryColor: '#7C3AED'  // Purple
primaryColor: '#EA580C'  // Orange
```

### Color Usage Map

The primary color appears in:
- ✅ Logo background (if no custom logo)
- ✅ Active navigation items
- ✅ Primary buttons
- ✅ Progress bars
- ✅ Badge backgrounds
- ✅ Link hover states
- ✅ Status indicators

---

## 📋 Implementation Checklist

When setting up white-label branding:

- [ ] Choose company name (official business name)
- [ ] Create company initials (2-4 letters)
- [ ] Decide on "Powered by AIO Fund" attribution
- [ ] (Optional) Select custom accent color
- [ ] (Optional) Prepare custom logo image (64x64px minimum)
- [ ] Update `/config/branding.ts` file
- [ ] Test on login page
- [ ] Test in sidebar navigation
- [ ] Test in guest learning center
- [ ] Clear browser cache and verify
- [ ] Document configuration for team reference

---

## 🔧 Advanced Configurations

### Multi-Brand Setup (Multiple Clients)

If you need to support multiple brands from a single codebase:

```typescript
// Create brand profiles
const brands = {
  nexxess: {
    companyName: 'Nexxess Business Advisors',
    companyInitials: 'NBA',
    showPoweredBy: true,
  },
  wealth: {
    companyName: 'Wealth Management Group',
    companyInitials: 'WMG',
    showPoweredBy: true,
    primaryColor: '#1E40AF',
  },
};

// Select brand based on subdomain or environment variable
const currentBrand = process.env.BRAND || 'nexxess';
export const brandingConfig = brands[currentBrand];
```

### Environment-Based Branding

```typescript
const isDevelopment = process.env.NODE_ENV === 'development';

export const brandingConfig: BrandingConfig = {
  companyName: isDevelopment ? 'AIO Fund (Dev)' : 'Nexxess Business Advisors',
  companyInitials: isDevelopment ? 'DEV' : 'NBA',
  showPoweredBy: true,
};
```

---

## 📐 Logo Specifications

### Recommended Logo Dimensions

| Context | Size | Format | Notes |
|---------|------|--------|-------|
| Sidebar | 48x48px | PNG/SVG | Square, transparent background |
| Login Page | 64x64px | PNG/SVG | Larger for prominence |
| Favicon | 32x32px | ICO/PNG | Browser tab icon |
| Email | 200x200px | PNG | High-res for communications |

### Logo Design Tips

1. **Keep it Simple:** Works at small sizes (48x48px)
2. **High Contrast:** Visible against dark (#0B1930) background
3. **Square Format:** Fits best in circular or square containers
4. **Transparent Background:** PNG with alpha channel preferred
5. **Vector Format:** SVG ideal for crisp scaling

---

## 🎯 Brand Identity Matrix

| Element | Location | Customizable | Notes |
|---------|----------|--------------|-------|
| Logo | Sidebar, Login | ✅ Yes | Via `logoUrl` or initials |
| Company Name | Sidebar, Login, Headers | ✅ Yes | `companyName` field |
| Primary Color | Buttons, Links, Badges | ✅ Yes | `primaryColor` field |
| "Powered By" Badge | Sidebar, Login | ✅ Yes | `showPoweredBy` toggle |
| Typography | All text | ❌ No | Fixed (Inter font) |
| Secondary Colors | Backgrounds, borders | ❌ No | Part of design system |

---

## 📞 Common Questions

### Q: Can I remove "Powered by AIO Fund" completely?

**A:** Yes, set `showPoweredBy: false` in your configuration. This requires appropriate licensing.

### Q: Can I use my company logo instead of initials?

**A:** Yes, provide the path to your logo image in the `logoUrl` field.

### Q: Does changing the primary color affect everything?

**A:** It affects accent elements (buttons, links, badges) but not the core UI structure or backgrounds.

### Q: Can I use a rectangular logo?

**A:** Yes, but square logos work best. Rectangular logos may be cropped or scaled.

### Q: How do I test different configurations?

**A:** Change the values in `/config/branding.ts`, save, and refresh your browser.

---

## 📄 License Note

White-labeling is available under AIO Fund partnership agreements. The "Powered by AIO Fund" attribution helps maintain the platform relationship while allowing your brand to be front and center.

For questions about licensing or custom branding needs, contact your AIO Fund account manager.
