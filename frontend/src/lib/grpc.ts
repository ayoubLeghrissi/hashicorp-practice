import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import path from 'path';

// Proto files paths
const PROTO_DIR = path.join(process.cwd(), 'proto');
const PRODUCT_PROTO_PATH = path.join(PROTO_DIR, 'product', 'product.proto');
const INVENTORY_PROTO_PATH = path.join(PROTO_DIR, 'inventory', 'inventory.proto');
const CART_PROTO_PATH = path.join(PROTO_DIR, 'cart', 'cart.proto');

const loaderOptions: protoLoader.Options = {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
  includeDirs: [PROTO_DIR],
};

// Load product definition
const productPackageDefinition = protoLoader.loadSync(PRODUCT_PROTO_PATH, loaderOptions);
const productProto = grpc.loadPackageDefinition(productPackageDefinition) as any;

// Load inventory definition
const inventoryPackageDefinition = protoLoader.loadSync(INVENTORY_PROTO_PATH, loaderOptions);
const inventoryProto = grpc.loadPackageDefinition(inventoryPackageDefinition) as any;

// Load cart definition
const cartPackageDefinition = protoLoader.loadSync(CART_PROTO_PATH, loaderOptions);
const cartProto = grpc.loadPackageDefinition(cartPackageDefinition) as any;

// Get service addresses from environment (fallback to localhost for dev)
const PRODUCT_URL = process.env.PRODUCT_SERVICE_URL || 'localhost:50051';
const INVENTORY_URL = process.env.INVENTORY_SERVICE_URL || 'localhost:50052';
const CART_URL = process.env.CART_SERVICE_URL || 'localhost:50053';

// We create insecure credentials for development. 
// In production, you would configure TLS credentials here.
const credentials = grpc.credentials.createInsecure();

// Initialize clients
export const productClient = new productProto.product.ProductService(PRODUCT_URL, credentials);
export const inventoryClient = new inventoryProto.inventory.InventoryService(INVENTORY_URL, credentials);
export const cartClient = new cartProto.cart.CartService(CART_URL, credentials);
