function convertBRLToAnyCurrency(conversionRate, currencyName) {
    // Select all span elements on the page
    const spans = document.querySelectorAll('span');

    // Iterate over each span element
    spans.forEach(span => {
        // Extract the text content
        let text = span.textContent;

        // Check if the text matches the BRL format (e.g., R$9,205.04)
        let match = text.match(/R\$\s*([\d,]+.\d{2})/);

        if (match) {
            // Convert BRL amount string to a float (handling both commas and periods)
            let brlAmount = parseFloat(match[1].replace(',', ''));

            // Convert BRL to CAD
            let cadAmount = (brlAmount * conversionRate).toFixed(2);

            // Format the CAD amount with commas and add the CA$ symbol
            let cadFormatted = new Intl.NumberFormat('en-' + currencyName, { style: 'currency', currency: currencyName }).format(cadAmount);

            // Replace the text content with the new CAD value
            span.textContent = cadFormatted;
        }
    });
}