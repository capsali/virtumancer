# SPICE Console Responsive Scaling

## Overview

The SPICE console in VirtuMancer has been enhanced with responsive scaling to ensure that VM console windows fit properly within the browser viewport, regardless of screen size or VM resolution.

## Implementation

### Files Modified/Created

1. **`spice_responsive.css`** - Custom CSS that overrides default SPICE styling
2. **`spice_responsive.html`** - Enhanced SPICE client with responsive behavior
3. **`SpiceView.vue`** - Updated Vue component with improved container handling

### Key Features

#### 1. Responsive Scaling
- Console automatically scales to fit the browser window
- Maintains aspect ratio of the VM display
- Works on desktop, tablet, and mobile devices

#### 2. Smart Aspect Ratio Handling
- Calculates optimal scaling based on container vs. canvas dimensions
- Fits to width or height depending on which constraint is more restrictive
- Centers the console within the available space

#### 3. Dynamic Resize Support
- Responds to browser window resize events
- Automatically adjusts scaling when window size changes
- Triggers proper SPICE resize handling

#### 4. Enhanced User Experience
- Eliminates horizontal/vertical scrollbars on the console
- Provides full-screen console experience
- Maintains console quality with proper image rendering

### Technical Details

#### CSS Scaling Strategy
```css
.spice-screen canvas {
    max-width: 100% !important;
    max-height: 100% !important;
    width: auto !important;
    height: auto !important;
    object-fit: contain;
    box-sizing: border-box;
}

/* Overflow prevention */
* { max-width: 100% !important; }
body, html { overflow-x: hidden !important; }
```

#### JavaScript Scaling Logic
The `ensureProperScaling()` function:
1. Gets natural canvas dimensions
2. Calculates container aspect ratio
3. Determines optimal display size
4. Applies calculated dimensions while respecting CSS constraints

#### Vue Component Integration
- Uses `ref` to access iframe element
- Handles resize events
- Manages SPICE client status updates
- Provides proper cleanup on component unmount

### Browser Compatibility

- **Chrome/Edge**: Full support with hardware acceleration
- **Firefox**: Full support with proper image rendering
- **Safari**: Full support with webkit optimizations
- **Mobile browsers**: Responsive scaling with touch support

### Performance Considerations

- Uses CSS transforms for smooth scaling
- Leverages hardware acceleration when available
- Minimizes DOM reflows during resize operations
- Implements periodic scaling checks for robustness

### Configuration

The responsive SPICE client is automatically used when accessing VM consoles through:
- VM Detail View → Console button
- VM List View → Console button  
- VM Card → Console button

### Troubleshooting

If console scaling issues persist:

1. **Check browser zoom level** - Ensure browser is at 100% zoom
2. **Verify WebSocket connection** - Console requires active WebSocket to SPICE proxy
3. **Check VM resolution** - Very high VM resolutions may need additional scaling time
4. **Review browser console** - Look for SPICE connection errors or JavaScript issues

### File Transfer Feature

The SPICE console includes a modern file transfer system:

#### **Enhanced File Transfer UI**
- **Header button**: "Transfer Files" button in console header
- **Modal interface**: Clean drag-and-drop modal with browse functionality
- **File management**: Add/remove files before transfer
- **Transfer status**: Real-time feedback on transfer progress

#### **Usage**
1. Click "Transfer Files" button in console header
2. Drag files to modal or use "Browse Files" button
3. Review selected files and click "Transfer Files"
4. Files are sent to VM via SPICE protocol

#### **Technical Implementation**
- Hidden SPICE transfer area maintains protocol compatibility
- PostMessage communication between modal and SPICE client
- Programmatic file drop simulation for seamless integration

### Future Enhancements

Potential improvements for the SPICE console experience:

1. **Fullscreen mode** - Dedicated fullscreen toggle button
2. **Scaling presets** - 50%, 75%, 100%, 125% scaling options  
3. **Quality settings** - Adaptive quality based on connection speed
4. **Multi-monitor support** - Handle VMs with multiple displays
5. **Touch gestures** - Pinch-to-zoom and pan gestures on mobile devices
6. **File transfer progress** - Individual file progress indicators
7. **Transfer history** - Log of recent file transfers

## Related Files

- `web/src/views/SpiceView.vue` - Main console view component
- `web/public/spice/spice_responsive.html` - Enhanced SPICE client
- `web/public/spice/spice_responsive.css` - Responsive styling
- `web/src/router/index.ts` - SPICE console routing configuration