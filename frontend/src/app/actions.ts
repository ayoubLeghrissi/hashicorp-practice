'use server';

import { productClient, inventoryClient, cartClient } from '@/lib/grpc';

// Promisify gRPC calls for easier consumption in Server Actions

export async function fetchProducts(page = 1, pageSize = 20, category = '') {
    return new Promise((resolve, reject) => {
        productClient.ListProducts({ page, page_size: pageSize, category }, (err: any, response: any) => {
            if (err) return reject(err);
            resolve(response);
        });
    });
}

export async function fetchProduct(id: string) {
    return new Promise((resolve, reject) => {
        productClient.GetProduct({ id }, (err: any, response: any) => {
            if (err) return reject(err);
            resolve(response.product);
        });
    });
}

export async function checkStock(productId: string) {
    return new Promise((resolve, reject) => {
        inventoryClient.CheckStock({ product_id: productId, requested_quantity: 1 }, (err: any, response: any) => {
            if (err) return reject(err);
            resolve(response);
        });
    });
}

export async function fetchCart(userId: string) {
    return new Promise((resolve, reject) => {
        cartClient.GetCart({ user_id: userId }, (err: any, response: any) => {
            if (err) return reject(err);
            resolve(response.cart);
        });
    });
}

export async function addToCart(userId: string, productId: string, quantity: number = 1) {
    return new Promise((resolve, reject) => {
        cartClient.AddToCart({ user_id: userId, product_id: productId, quantity }, (err: any, response: any) => {
            if (err) return reject(err.details || err.message);
            if (!response.success) return reject(new Error(response.message || 'Failed to add to cart'));
            resolve(response.cart);
        });
    });
}

export async function removeFromCart(userId: string, productId: string) {
    return new Promise((resolve, reject) => {
        cartClient.RemoveFromCart({ user_id: userId, product_id: productId }, (err: any, response: any) => {
            if (err) return reject(err.details || err.message);
            if (!response.success) return reject(new Error('Failed to remove item from cart'));
            resolve(response.cart);
        });
    });
}

export async function checkoutCart(userId: string) {
    return new Promise((resolve, reject) => {
        cartClient.Checkout({ user_id: userId }, (err: any, response: any) => {
            if (err) return reject(err.details || err.message);
            if (!response.success) return reject(new Error(response.message || 'Checkout failed'));
            resolve(response);
        });
    });
}
