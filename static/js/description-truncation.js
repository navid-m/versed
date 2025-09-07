/**
 * Description Truncation Utility
 * Removes <p> tags beyond the 3rd one in description containers
 */

class DescriptionTruncator {
    constructor() {
        this.maxParagraphs = 3;
    }

    /**
     * Apply truncation to all description elements on the page
     */
    applyTruncation() {
        // Find all description containers with the specific class structure
        const descriptionContainers = document.querySelectorAll('.text-gray-700.dark\\:text-gray-300.text-sm.leading-relaxed.line-height-6.font-medium.tracking-wide.line-clamp-3.bg-gradient-to-br.from-gray-50\\/80.to-white\\/50.dark\\:from-gray-800\\/60.dark\\:to-gray-700\\/40.backdrop-blur-sm.rounded-lg.px-4.py-3.border-l-4.border-blue-500\\/30.dark\\:border-blue-400\\/40.shadow-sm');

        descriptionContainers.forEach(container => {
            this.truncateContainer(container);
        });
    }

    /**
     * Truncate a single description container
     * @param {Element} container - The description container element
     */
    truncateContainer(container) {
        const paragraphs = container.querySelectorAll('p');

        if (paragraphs.length <= this.maxParagraphs) {
            return; // No truncation needed
        }

        // Remove paragraphs beyond the 3rd one
        for (let i = this.maxParagraphs; i < paragraphs.length; i++) {
            paragraphs[i].remove();
        }

        // Add ellipsis to the last remaining paragraph if it doesn't already have content
        const lastParagraph = container.querySelector('p:last-child');
        if (lastParagraph && (!lastParagraph.textContent || lastParagraph.textContent.trim() === '')) {
            lastParagraph.textContent = '...';
        } else {
            // Add a new paragraph with ellipsis
            const ellipsisPara = document.createElement('p');
            ellipsisPara.textContent = '...';
            container.appendChild(ellipsisPara);
        }
    }
}

// Initialize truncation when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    const truncator = new DescriptionTruncator();
    truncator.applyTruncation();

    // Also apply truncation after new content is loaded (for dynamic content)
    const observer = new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
            if (mutation.type === 'childList') {
                mutation.addedNodes.forEach((node) => {
                    if (node.nodeType === Node.ELEMENT_NODE) {
                        // Check if the added node contains description containers
                        const containers = node.querySelectorAll ?
                            node.querySelectorAll('.text-gray-700.dark\\:text-gray-300.text-sm.leading-relaxed.line-height-6.font-medium.tracking-wide.line-clamp-3.bg-gradient-to-br.from-gray-50\\/80.to-white\\/50.dark\\:from-gray-800\\/60.dark\\:to-gray-700\\/40.backdrop-blur-sm.rounded-lg.px-4.py-3.border-l-4.border-blue-500\\/30.dark\\:border-blue-400\\/40.shadow-sm') :
                            [];

                        // Also check if the node itself is a description container
                        if (node.matches && node.matches('.text-gray-700.dark\\:text-gray-300.text-sm.leading-relaxed.line-height-6.font-medium.tracking-wide.line-clamp-3.bg-gradient-to-br.from-gray-50\\/80.to-white\\/50.dark\\:from-gray-800\\/60.dark\\:to-gray-700\\/40.backdrop-blur-sm.rounded-lg.px-4.py-3.border-l-4.border-blue-500\\/30.dark\\:border-blue-400\\/40.shadow-sm')) {
                            truncator.truncateContainer(node);
                        }

                        // Apply to found containers
                        containers.forEach(container => {
                            truncator.truncateContainer(container);
                        });
                    }
                });
            }
        });
    });

    observer.observe(document.body, {
        childList: true,
        subtree: true
    });
});

// Export for use in other modules
window.DescriptionTruncator = DescriptionTruncator;
