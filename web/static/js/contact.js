/**
 * Contact Form — captcha generation, validation, and submission.
 * Messages use the global toast system.
 */

import { showToast } from './modules/toast.js';

let ac = null;

export function init() {
    ac = new AbortController();
    const { signal } = ac;
    const contactForm = document.getElementById('contactForm');
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
            showToast('Please fill in all required fields.', 'error');
            return;
        }

        if (parseInt(captchaValue) !== captchaAnswer) {
            showToast('Please solve the math problem correctly.', 'error');
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
                    showToast(result.message || 'Message sent! You\'ll receive a response soon.', 'success');
                    contactForm.reset();
                } else {
                    showToast(result.message || 'Failed to send message. Please try again.', 'error');
                }
                generateCaptcha();
            })
            .catch((error) => {
                console.error('Contact form error:', error);
                showToast('An error occurred. Please try again later.', 'error');
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
