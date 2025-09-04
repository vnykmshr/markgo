/* Contact Page JavaScript */

document.addEventListener("DOMContentLoaded", function () {
  const contactForm = document.getElementById("contactForm");
  const messageContainer = document.getElementById("messageContainer");
  const messageTextarea = document.getElementById("message");
  const charCount = document.getElementById("charCount");
  const captchaInput = document.getElementById("captcha");
  const captchaQuestionField = document.getElementById("captchaQuestionField");
  const num1Element = document.getElementById("num1");
  const num2Element = document.getElementById("num2");
  const refreshBtn = document.getElementById("refreshCaptcha");

  let captchaAnswer = 0;

  // Generate simple addition captcha
  function generateCaptcha() {
    const num1 = Math.floor(Math.random() * 10) + 1;
    const num2 = Math.floor(Math.random() * 10) + 1;
    captchaAnswer = num1 + num2;

    num1Element.textContent = num1;
    num2Element.textContent = num2;
    captchaQuestionField.value = `${num1} + ${num2}`;
    captchaInput.value = "";
  }

  // Character counter
  function updateCharCount() {
    const count = messageTextarea.value.length;
    charCount.textContent = count;
    charCount.style.color =
      count > 2000 ? "#dc3545" : count > 1800 ? "#fd7e14" : "#6c757d";
  }

  // Show message
  function showMessage(message, type = "success") {
    const icon =
      type === "success"
        ? '<path d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zm-3.97-3.03a.75.75 0 0 0-1.08.022L7.477 9.417 5.384 7.323a.75.75 0 0 0-1.06 1.061L6.97 11.03a.75.75 0 0 0 1.079-.02l3.992-4.99a.75.75 0 0 0-.01-1.05z"/>'
        : '<path d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zM5.354 4.646a.5.5 0 1 0-.708.708L7.293 8l-2.647 2.646a.5.5 0 0 0 .708.708L8 8.707l2.646 2.647a.5.5 0 0 0 .708-.708L8.707 8l2.647-2.646a.5.5 0 0 0-.708-.708L8 7.293 5.354 4.646z"/>';

    messageContainer.innerHTML = `
      <div class="message message-${type}">
        <div class="message-icon">
          <svg width="20" height="20" fill="currentColor" viewBox="0 0 16 16">${icon}</svg>
        </div>
        <div class="message-content">
          <p>${message}</p>
        </div>
        <button class="message-close" onclick="this.parentElement.remove()">×</button>
      </div>
    `;

    if (type === "success") {
      setTimeout(() => {
        const messageEl = messageContainer.querySelector(".message");
        if (messageEl) messageEl.remove();
      }, 5000);
    }

    messageContainer.scrollIntoView({ behavior: "smooth" });
  }

  // Form submission
  function handleSubmit(e) {
    e.preventDefault();

    const submitBtn = contactForm.querySelector(".btn-submit");
    const formData = new FormData(contactForm);

    // Basic validation
    const name = formData.get("name")?.trim();
    const email = formData.get("email")?.trim();
    const subject = formData.get("subject")?.trim();
    const message = formData.get("message")?.trim();
    const captchaValue = formData.get("captcha_answer")?.trim();

    if (!name || !email || !subject || !message || !captchaValue) {
      showMessage("Please fill in all required fields.", "error");
      return;
    }

    // Validate captcha
    if (parseInt(captchaValue) !== captchaAnswer) {
      showMessage("Please solve the math problem correctly.", "error");
      generateCaptcha();
      return;
    }

    // Show loading state
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span class="spinner"></span> Sending...';

    // Submit form
    const data = {
      name,
      email,
      subject,
      message,
      captcha_question: formData.get("captcha_question"),
      captcha_answer: captchaValue,
    };

    fetch("/contact", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    })
      .then((response) => response.json())
      .then((result) => {
        if (result.success) {
          showMessage(
            result.message ||
              "Thank you for your message! I'll get back to you soon.",
            "success",
          );
          contactForm.reset();
          updateCharCount();
          generateCaptcha();
        } else {
          showMessage(
            result.message || "Failed to send message. Please try again.",
            "error",
          );
          generateCaptcha();
        }
      })
      .catch((error) => {
        console.error("Contact form error:", error);
        showMessage("An error occurred. Please try again later.", "error");
        generateCaptcha();
      })
      .finally(() => {
        submitBtn.disabled = false;
        submitBtn.innerHTML = `
        <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16">
          <path d="M15.854.146a.5.5 0 0 1 .11.54L13.026 8.74a.5.5 0 0 1-.708.251L9 7.5 7.5 9l-1.49-3.318a.5.5 0 0 1 .251-.708L14.31 2.036a.5.5 0 0 1 .54.11z"/>
          <path d="M2.5 3a.5.5 0 0 0-.5.5v9a.5.5 0 0 0 .5.5h11a.5.5 0 0 0 .5-.5V6.5a.5.5 0 0 0-1 0V12H3V4h5.5a.5.5 0 0 0 0-1H2.5z"/>
        </svg>
        Send Message
      `;
      });
  }

  // Event listeners
  refreshBtn.addEventListener("click", generateCaptcha);
  contactForm.addEventListener("submit", handleSubmit);
  messageTextarea.addEventListener("input", updateCharCount);

  contactForm.addEventListener("reset", () => {
    setTimeout(() => {
      updateCharCount();
      generateCaptcha();
    }, 0);
  });

  // Initialize
  generateCaptcha();
  updateCharCount();
});
