# What is the Service Registration & Discovery

- Service Registration & discovery là 1 thành phần quan trọng trong kiến trúc microservices, nó cho phép các dịch vụ trong hệ thống phát hiện và giao tiếp với nhau một cách động mà không cần phải cấu hình thủ công các điểm cuối của dịch vụ

- Service Registration: khi 1 dịch vụ khởi động, nó sẽ đang ký bản thân với 1 trung tâm đăng kí dịch vụ (service registry). Thông tin đăng kí bao gồm địa chỉ IP, cổng, và các thông tin khác cần thiết để truy cập dịch vụ

- Service