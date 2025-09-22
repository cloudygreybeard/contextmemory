# Chat Export Test

**User**: How do I implement authentication in my Node.js API?

**Assistant**: I'll help you implement authentication in your Node.js API. Here are the key approaches:

## JWT Authentication

```javascript
const jwt = require('jsonwebtoken');

function generateToken(user) {
    return jwt.sign({
        userId: user.id,
        email: user.email
    }, process.env.JWT_SECRET, { expiresIn: '7d' });
}
```

## Middleware for Route Protection

```javascript
function authenticateToken(req, res, next) {
    const token = req.headers.authorization?.split(' ')[1];
    
    if (!token) {
        return res.status(401).json({ error: 'Access denied' });
    }
    
    jwt.verify(token, process.env.JWT_SECRET, (err, user) => {
        if (err) return res.status(403).json({ error: 'Invalid token' });
        req.user = user;
        next();
    });
}
```

This approach provides secure, stateless authentication perfect for APIs.

**User**: What about password hashing? How should I handle that?

**Assistant**: Great question! Password security is crucial. Here's the proper approach:

```javascript
const bcrypt = require('bcrypt');

// Hash password before saving
async function hashPassword(password) {
    const saltRounds = 12;
    return await bcrypt.hash(password, saltRounds);
}

// Verify password during login
async function verifyPassword(password, hashedPassword) {
    return await bcrypt.compare(password, hashedPassword);
}
```

Never store plain text passwords - always use bcrypt with a salt rounds of 10-12.