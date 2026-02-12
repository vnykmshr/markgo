/**
 * Contact Form — captcha generation, validation, and submission.
 */

let ac = null;

export function init() {
    ac = new AbortController();
    const { signal } = ac;
    const contactForm = document.getElementById('contactForm');
    const messageContainer = document.getElementById('messageContainer');
    const captchaInput = document.getElementById('captcha');
    const captchaQuestionField = document.getElementById('captchaQuestionField');
    const num1Element = document.getElementById('num1');
    const num2Element = document.getElementById('num2');
    const refreshBtn = document.getElementById('refreshCaptcha');

    if (!contactForm) return;

    let captchaAnswer = 0;

    function generateCaptcha() {
        const num1 = Math.floor(Math.random() * 10) + 1;
        const num2 = Math.floor(Math.random() * 10) + 1;
        captchaAnswer = num1 + num2;

        num1Element.textContent = num1;
        num2Element.textContent = num2;
        captchaQuestionField.value = `${num1} + ${num2}`;
        captchaInput.value = '';
    }

    function showMessage(message, type = 'success') {
        const div = document.createElement('div');
        div.className = `message message-${type}`;

        const iconSvg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        iconSvg.setAttribute('width', '20');
        iconSvg.setAttribute('height', '20');
        iconSvg.setAttribute('fill', 'currentColor');
        iconSvg.setAttribute('viewBox', '0 0 16 16');
        const iconPath = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        iconPath.setAttribute('d', type === 'success'
            ? 'M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zm-3.97-3.03a.75.75 0 0 0-1.08.022L7.477 9.417 5.384 7.323a.75.75 0 0 0-1.06 1.061L6.97 11.03a.75.75 0 0 0 1.079-.02l3.992-4.99a.75.75 0 0 0-.01-1.05z'
            : 'M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zM5.354 4.646a.5.5 0 1 0-.708.708L7.293 8l-2.647 2.646a.5.5 0 0 0 .708.708L8 8.707l2.646 2.647a.5.5 0 0 0 .708-.708L8.707 8l2.647-2.646a.5.5 0 0 0-.708-.708L8 7.293 5.354 4.646z');
        iconSvg.appendChild(iconPath);

        const iconDiv = document.createElement('div');
        iconDiv.className = 'message-icon';
        iconDiv.appendChild(iconSvg);
        div.appendChild(iconDiv);

        const contentDiv = document.createElement('div');
        contentDiv.className = 'message-content';
        const p = document.createElement('p');
        p.textContent = message;
        contentDiv.appendChild(p);
        div.appendChild(contentDiv);

        const closeBtn = document.createElement('button');
        closeBtn.className = 'message-close';
        closeBtn.textContent = '\u00d7';
        closeBtn.addEventListener('click', () => div.remove());
        div.appendChild(closeBtn);

        messageContainer.textContent = '';
        messageContainer.appendChild(div);

        if (type === 'success') {
            setTimeout(() => { if (div.parentNode) div.remove(); }, 5000);
        }

        messageContainer.scrollIntoView({ behavior: 'smooth' });
    }

    function handleSubmit(e) {
        e.preventDefault();

        const submitBtn = contactForm.querySelector('.contact-submit');
        const formData = new FormData(contactForm);

        const name = (formData.get('name') || '').trim();
        const email = (formData.get('email') || '').trim();
        const subject = (formData.get('subject') || '').trim();
        const message = (formData.get('message') || '').trim();
        const captchaValue = (formData.get('captcha_answer') || '').trim();

        if (!name || !email || !subject || !message || !captchaValue) {
            showMessage('Please fill in all required fields.', 'error');
            return;
        }

        if (parseInt(captchaValue) !== captchaAnswer) {
            showMessage('Please solve the math problem correctly.', 'error');
            generateCaptcha();
            return;
        }

        submitBtn.disabled = true;
        const spinner = document.createElement('span');
        spinner.className = 'spinner';
        submitBtn.textContent = ' Sending...';
        submitBtn.prepend(spinner);

        fetch('/contact', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, email, subject, message, captcha_question: formData.get('captcha_question'), captcha_answer: captchaValue }),
        })
            .then((response) => response.json())
            .then((result) => {
                if (result.success) {
                    showMessage(result.message || 'Thank you for your message! You\'ll receive a response soon.', 'success');
                    contactForm.reset();
                } else {
                    showMessage(result.message || 'Failed to send message. Please try again.', 'error');
                }
                generateCaptcha();
            })
            .catch((error) => {
                console.error('Contact form error:', error);
                showMessage('An error occurred. Please try again later.', 'error');
                generateCaptcha();
            })
            .finally(() => {
                submitBtn.disabled = false;
                submitBtn.textContent = 'Send';
            });
    }

    refreshBtn.addEventListener('click', generateCaptcha, { signal });
    contactForm.addEventListener('submit', handleSubmit, { signal });
    contactForm.addEventListener('reset', () => setTimeout(generateCaptcha, 0), { signal });

    generateCaptcha();
}

export function destroy() {
    if (ac) { ac.abort(); ac = null; }
}
