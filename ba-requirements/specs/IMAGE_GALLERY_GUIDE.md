# Image Gallery Feature Guide

## Overview
The AIO Fund Client Dashboard now includes a full-featured image gallery/album viewer that allows users to click on any image to view it in a full-screen lightbox with navigation, zoom, and download capabilities.

## Features

### 1. **Full-Screen Lightbox View**
- Click any image to open it in an immersive full-screen viewer
- Dark background (black/95% opacity) for optimal image viewing
- Clean, distraction-free interface

### 2. **Navigation Controls**
- **Left/Right Arrows**: Navigate between images in the gallery
- **Keyboard Support**: 
  - `←` Left arrow key - Previous image
  - `→` Right arrow key - Next image
  - `Esc` - Close gallery
- **Large Navigation Buttons**: Circular buttons on left/right sides for easy clicking
- **Image Counter**: Shows current position (e.g., "3 / 12")

### 3. **Zoom Controls**
- **Zoom In**: Magnify images up to 300% (3x)
- **Zoom Out**: Reduce to 50% (0.5x)
- **Zoom Percentage Display**: Shows current zoom level
- Zoom resets when navigating to a different image

### 4. **Download Functionality**
- Download button to save images locally
- Preserves original image quality

### 5. **Image Information**
- Displays image title at the top
- Shows image description for context
- Gradient overlays for better text readability

## Where It's Implemented

### 1. **Dashboard - Recent Documents**
- Location: Dashboard page, "Recent Documents" section
- Purpose: View uploaded trust documents and verification files
- 4 recent documents displayed in a grid
- Hover effect reveals document title

### 2. **Trust Detail Page - Uploaded Documents & Photos**
- Location: My Trust → Select a trust → "Uploaded Documents & Photos" section
- Purpose: View all documents related to a specific trust
- Includes TSA documents, ID verification, proof of address, beneficiary info
- Grid layout with 2-4 columns (responsive)

### 3. **Marketplace - Fund Detail Images**
- Location: Marketplace → Select a fund → Image gallery at top
- Purpose: View fund-related imagery and visual representations
- 6 images per fund (except one reserved for manager info)
- Click any image except the manager info card to view in gallery

## User Interaction Flow

### Opening the Gallery
1. Hover over any image in the grid
2. Image border changes to gold (#C6A661)
3. Image scales up slightly (110%)
4. Dark gradient overlay appears with title
5. Small icon appears in top-right corner
6. Click anywhere on the image to open gallery

### Navigating Images
1. Use large circular buttons on sides
2. Use arrow buttons in the bottom control bar
3. Use keyboard arrow keys
4. Swipe gestures (on touch devices)

### Zoom Controls
1. Click zoom in (+) or zoom out (-) buttons
2. Current zoom percentage displays in center
3. Zoom is limited between 50% and 300%
4. Zoom automatically resets when changing images

### Closing the Gallery
1. Click the X button in top-right corner
2. Press Escape key
3. Click outside the image area (on the dark background)

## Technical Implementation

### Component Structure
```
ImageGalleryModal
├── Header (title, description, close button)
├── Image Display Area (with zoom transformation)
├── Controls Bar
│   ├── Navigation (prev/next, counter)
│   ├── Zoom Controls (zoom in/out, percentage)
│   └── Download Button
└── Large Navigation Arrows (left/right sides)
```

### Key Features
- **Responsive Design**: Works on desktop, tablet, and mobile
- **Keyboard Accessibility**: Full keyboard navigation support
- **State Management**: Tracks current image index and zoom level
- **Automatic Reset**: Zoom resets when changing images
- **Circular Navigation**: Wraps from last to first image and vice versa

## Customization Options

To add an image gallery to a new page:

1. Import the required components:
```tsx
import { ImageWithFallback } from '../figma/ImageWithFallback';
import { ImageGalleryModal } from '../ImageGalleryModal';
```

2. Set up state:
```tsx
const [isGalleryOpen, setIsGalleryOpen] = useState(false);
const [selectedImageIndex, setSelectedImageIndex] = useState(0);
```

3. Create image data:
```tsx
const images = [
  {
    url: 'image-url.jpg',
    title: 'Image Title',
    description: 'Image description',
  },
  // ... more images
];
```

4. Create clickable image grid:
```tsx
<div className="grid grid-cols-4 gap-4">
  {images.map((image, index) => (
    <div
      key={index}
      onClick={() => {
        setSelectedImageIndex(index);
        setIsGalleryOpen(true);
      }}
      className="cursor-pointer hover:opacity-80"
    >
      <ImageWithFallback src={image.url} alt={image.title} />
    </div>
  ))}
</div>
```

5. Add the modal:
```tsx
<ImageGalleryModal
  images={images}
  isOpen={isGalleryOpen}
  onClose={() => setIsGalleryOpen(false)}
  initialIndex={selectedImageIndex}
/>
```

## Design System Integration

The image gallery follows the AIO Fund design system:

- **Colors**: 
  - Background: Black at 95% opacity
  - Accent: Gold (#C6A661) for hover states
  - Controls: White with transparency
  
- **Typography**: 
  - Uses Inter font family
  - Title: White color for visibility
  - Description: Gray-300 for secondary text

- **Border Radius**: 10px for cards, rounded-full for buttons

- **Shadows**: Subtle shadows on hover for depth

## Best Practices

1. **Image Quality**: Use high-resolution images (at least 800px width)
2. **Alt Text**: Always provide descriptive alt text for accessibility
3. **Loading States**: The ImageWithFallback component handles loading states
4. **Error Handling**: Fallback image shown if image fails to load
5. **Performance**: Images are lazy-loaded and optimized
6. **Mobile**: Touch-friendly button sizes (min 44x44px)

## Future Enhancements

Potential improvements for future versions:
- Pinch-to-zoom on mobile devices
- Image rotation controls
- Slideshow/autoplay mode
- Thumbnail strip at bottom
- Social sharing capabilities
- Comparison view (side-by-side)
- Annotations and markup tools
