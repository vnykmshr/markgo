/* 404 Error Page JavaScript */

document.addEventListener("DOMContentLoaded", function () {
    // Easter egg trigger - click on the 404 text multiple times
    const illustration = document.querySelector(
        ".error-illustration svg text",
    );
    const easterEgg = document.getElementById("easterEgg");
    let clickCount = 0;

    if (illustration && easterEgg) {
        illustration.style.cursor = "pointer";

        illustration.addEventListener("click", function () {
            clickCount++;

            // Add a little animation on each click
            this.style.transform = "scale(1.1)";
            setTimeout(() => {
                this.style.transform = "scale(1)";
            }, 150);

            // Show easter egg after 5 clicks
            if (clickCount >= 5) {
                easterEgg.style.display = "block";

                // Hide it after 10 seconds
                setTimeout(() => {
                    easterEgg.style.display = "none";
                    clickCount = 0; // Reset counter
                }, 10000);
            }
        });
    }

    // Auto-focus search input
    const searchInput = document.querySelector(".search-input");
    if (searchInput && window.innerWidth > 768) {
        setTimeout(() => {
            searchInput.focus();
        }, 1000);
    }

    // Add some random animation to floating elements
    const floatingElements = document.querySelectorAll(
        ".error-illustration svg circle, .error-illustration svg rect, .error-illustration svg polygon",
    );

    floatingElements.forEach((element, index) => {
        const randomDelay = Math.random() * 2;
        const randomDuration = 3 + Math.random() * 2;

        element.style.animation = `float ${randomDuration}s ease-in-out ${randomDelay}s infinite alternate`;
    });
});
