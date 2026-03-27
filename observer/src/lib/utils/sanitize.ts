import DOMPurify from 'isomorphic-dompurify';

export const sanitize = (html: string) => {
	// restrict to essential inline formatting tags for safety
	return DOMPurify.sanitize(html, {
		ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'span', 'br', 'p'],
		ALLOWED_ATTR: ['class'],
	});
};
