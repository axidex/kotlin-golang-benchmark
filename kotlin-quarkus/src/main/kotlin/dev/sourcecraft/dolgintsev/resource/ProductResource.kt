package dev.sourcecraft.dolgintsev.resource

import jakarta.transaction.Transactional
import jakarta.ws.rs.*
import jakarta.ws.rs.core.MediaType
import jakarta.ws.rs.core.Response
import dev.sourcecraft.dolgintsev.entity.Product
import org.jboss.logging.Logger

@Path("/api/products")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
class ProductResource {

    private val log = Logger.getLogger(ProductResource::class.java)

    @GET
    fun getAllProducts(): List<Product> {
        log.info("GET /api/products - fetching all products")
        val products = Product.listAll()
        log.info("GET /api/products - returned ${products.size} products")
        return products
    }

    @GET
    @Path("/{id}")
    fun getProductById(@PathParam("id") id: Long): Response {
        log.info("GET /api/products/$id - fetching product by id")
        val product = Product.findById(id)
        return if (product != null) {
            log.info("GET /api/products/$id - product found: ${product.name}")
            Response.ok(product).build()
        } else {
            log.warn("GET /api/products/$id - product not found")
            Response.status(Response.Status.NOT_FOUND).build()
        }
    }

    @POST
    @Transactional
    fun createProduct(product: Product): Response {
        log.info("POST /api/products - creating product: ${product.name}")
        product.persist()
        log.info("POST /api/products - product created with id: ${product.id}")
        return Response.status(Response.Status.CREATED).entity(product).build()
    }

    @PUT
    @Path("/{id}")
    @Transactional
    fun updateProduct(@PathParam("id") id: Long, updatedProduct: Product): Response {
        log.info("PUT /api/products/$id - updating product")
        val product = Product.findById(id)
        return if (product != null) {
            product.name = updatedProduct.name
            product.description = updatedProduct.description
            product.price = updatedProduct.price
            product.quantity = updatedProduct.quantity
            log.info("PUT /api/products/$id - product updated: ${product.name}")
            Response.ok(product).build()
        } else {
            log.warn("PUT /api/products/$id - product not found")
            Response.status(Response.Status.NOT_FOUND).build()
        }
    }

    @DELETE
    @Path("/{id}")
    @Transactional
    fun deleteProduct(@PathParam("id") id: Long): Response {
        log.info("DELETE /api/products/$id - deleting product")
        val deleted = Product.deleteById(id)
        return if (deleted) {
            log.info("DELETE /api/products/$id - product deleted successfully")
            Response.noContent().build()
        } else {
            log.warn("DELETE /api/products/$id - product not found")
            Response.status(Response.Status.NOT_FOUND).build()
        }
    }
}
